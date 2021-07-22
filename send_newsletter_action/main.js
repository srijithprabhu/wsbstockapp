const Cloudant = require("@cloudant/cloudant");
const Mustache = require("mustache");
const nodemailer = require("nodemailer");

const EMAIL_USER_POSTFIX = "+wallstreetbets";
const EMAIL_REGEX = /([^@]+)(@.+)/g;

function getAllSubredditThreads(params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const cloudant_dbname = params["dbname"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(cloudant_dbname);
    return db.list({include_docs: true}).then((response) => {
        let results = response.rows;
        return results.reduce((result, row) => {
            const doc = row.doc
            const subreddit = doc._id;
            const threads = doc.threads;
            threads.forEach((thread) => {
                thread.subreddit = subreddit;
                result.push(thread);
            })
            return result;
        }, []);
    });
}

function getEmailTransporter(params) {
    const username = params["email_address"]
    const password = params["email_password"]
    const server = params["email_smtp_host"]
    const port = params["email_smtp_port"]

    let transporter = nodemailer.createTransport({
        host: server,
        port: port,
        secure: false,
        auth: {
            user: username,
            pass: password,
        },
    });
    return transporter;
}

function generateToEmailAddress(user_email) {
    const matched_email = user_email.match(EMAIL_REGEX);
    return `${matched_email[1]}${EMAIL_USER_POSTFIX}${matched_email[2]}`;
}

function generateRedditThreadFilter(subreddit_filter_specs) {
    const subreddit_to_flair = subreddit_filter_specs.reduce((so_far, element) => {
        so_far[element.subreddit] = element.flair;
        return so_far;
    },{})
    return function(element) {
        if (Object.keys(subreddit_to_flair).indexOf(element.subreddit) > -1) {
            const flair = subreddit_to_flair[element.subreddit];
            if (!flair) {
                return true;
            } else {
                return flair.indexOf(element.link_flair) > -1;
            }
        }
        return false;
    }
}

const MUSTACHE_TEMPLATE = "Good Morning {{ name }}, here are the threads I found today:" +
    "<ul>" +
    "{{#threads}}" +
    "<li><a href={{url}}>{{title}}</a>(Shorts: {{shorts}})(Subreddit/Flair: {{subreddit}}/{{link_flair}})(Upvotes: {{upvotes}})</li>" +
    "{{/threads}}" +
    "</ul>";

function generateNewsletterBody(name, threads) {
    return Mustache.render(MUSTACHE_TEMPLATE, {name, threads});
}

function createAndSendNewsletters(threads, params) {
    const emailTransporter = getEmailTransporter(params);
    const users = params.users;
    const date = new Date();
    const sentEmails = users.map((user) => {
        const to_email = generateToEmailAddress(user.email);
        console.log(user.email);
        console.log(to_email);
        console.log(params['email_address']);
        const filter = generateRedditThreadFilter(user.subreddits);
        const newsletter_threads = threads.filter(filter);
        const newsletter_body = generateNewsletterBody(user.name, newsletter_threads);
        return emailTransporter.sendMail({
            from: params["email_address"], // sender address
            to: to_email, // list of receivers
            subject: `WallstreetBets Results for ${date.toDateString()}`, // Subject line
            html: newsletter_body, // html body
        });
    });
    return Promise.all(sentEmails);
}

function getHighShortInterestData(params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const document_id = params["short_interest_document_id"]
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(params["short_interest_dbname"]);
    return db.get(document_id).then((doc) => {
        return doc.data;
    });
}

function getListOfRegexes(high_interest_shorts) {
    return high_interest_shorts.map((high_interest_short) => {
       return new RegExp(`(?:${high_interest_short.symbol}|${high_interest_short.name})`, "i");
    });
}

function threadMatchesRegex(thread, regex) {
    return regex.test(thread.title) || regex.test(thread.text);
}

function populateShortsForRedditThreads(threads, shorts) {
    const regexes = getListOfRegexes(shorts)
    return threads.forEach((thread) => {
        thread.shorts = regexes.reduce((accumulator, regex, index) => {
            if (threadMatchesRegex(thread,regex)) {
                accumulator.push(shorts[index].symbol)
            }
            return accumulator;
        }, []).join(", ");
    })
}

function getAllSubredditThreadsAndShorts(params) {
    return Promise.all([
        getAllSubredditThreads(params),
        getHighShortInterestData(params)
    ]).then((results) => {
        const subreddit_threads = results[0];
        const high_interest_shorts = results[1];
        populateShortsForRedditThreads(subreddit_threads, high_interest_shorts);
        return subreddit_threads;
    })
}

function main(params) {
    return getAllSubredditThreadsAndShorts(params)
        .then((subreddit_threads) => {
            subreddit_threads.sort((a, b) => {
                return b.upvotes - a.upvotes;
            })
            return createAndSendNewsletters(subreddit_threads, params);
        })
}