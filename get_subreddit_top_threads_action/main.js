/**
 *
 * main() will be run when you invoke this action
 *
 * @param Cloud Functions actions accept a single parameter, which must be a JSON object.
 *
 * @return The output of this action, which must be a JSON object.
 *
 */
const RedditAccessTokenEndpoint = "https://www.reddit.com/api/v1/access_token"
const SubredditEndpointFormat = ""

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
                return response.body;
            });
        })
}

function getRedditAuthorization(params) {
    let client_id = params["reddit_client_id"]
    let secret_token = params["reddit_secret"]
    let username = params["reddit_username"]
    let password = params["reddit_password"]

    return axios.get(RedditAccessTokenEndpoint, {
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

function main(params) {
    return getRedditThreads(params);
}
