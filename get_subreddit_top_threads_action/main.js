/**
 *
 * main() will be run when you invoke this action
 *
 * @param Cloud Functions actions accept a single parameter, which must be a JSON object.
 *
 * @return The output of this action, which must be a JSON object.
 *
 */
const RedditAccessTokenEndpoint = "https://www.reddit.com/api/v1/access_token";

const Cloudant = require("@cloudant/cloudant");
const axios = require("axios").default;

function getRedditThreads(params) {
    let subreddit = params["subreddit"];
    let limit = params["limit"];
    let timespan = params["timespan"]
    return getRedditAuthorization(params)
        .then((body) => {
            return axios.get(`https://oauth.reddit.com/r/${subreddit}/top`, {
                params: {
                    limit: limit,
                    t: timespan
                },
                headers: {
                    "Authorization": `${body.token_type} ${body.access_token}`,
                    "User-Agent": "MyAuthBot/0.0.1"
                }
            }).then((response) => {
                return response.data;
            });
        })
}

function getRedditAuthorization(params) {
    let client_id = params["reddit_client_id"]
    let secret_token = params["reddit_secret"]
    let username = params["reddit_username"]
    let password = params["reddit_password"]

    return axios.post(RedditAccessTokenEndpoint, null,{
        params: {
            "grant_type": "password",
            username,
            password
        },
        headers: {
            "User-Agent": "MyAuthBot/0.0.1"
        },
        responseType: "json",
        auth: {
            username: client_id,
            password: secret_token
        }
    }).then((response) => {
        return response.data
    });
}

function uploadRedditThread(data, params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const document_id = params["subreddit"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use("reddit-threads");
    return db.get(document_id).catch((err) => {
        console.log(err);
    }).then((doc) => {
        if (!doc) {
            doc = {}
        }
        doc._id = document_id;
        doc.thread = data;
        return db.insert(doc);
    });

}

function main(params) {
    return getRedditThreads(params).then((data) => {
        return uploadRedditThread(data, params);
    });
}
