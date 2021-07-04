const Cloudant = require("@cloudant/cloudant");
const Mustache = require("mustache");
const nodemailer = require("nodemailer");

function getAllSubredditThreads(params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const cloudant_dbname = params["dbname"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(cloudant_dbname);
    return db.list();
}

function main(params) {
    return getAllSubredditThreads(params)
        .then((subreddit_threads_map) => {
            console.log(subreddit_threads_map);
        })
}