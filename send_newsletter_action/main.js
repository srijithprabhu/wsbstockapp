const Cloudant = require("@cloudant/cloudant");
const Mustache = require("mustache");
const nodemailer = require("nodemailer");

function getAllSubredditThreads(params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const cloudant_dbname = params["dbname"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(cloudant_dbname);
    return db.list({include_docs: true}).then((response) => {
        let results = response.rows;
        return results.reduce((result, row) => {
            const subreddit = row._id;
            const threads = row.threads;
            result[subreddit] = threads;
            return result;
        }, {});
    });
}

function main(params) {
    return getAllSubredditThreads(params)
        .then((subreddit_threads_map) => {
            return subreddit_threads_map;
        })
}