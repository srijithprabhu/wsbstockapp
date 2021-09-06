const Cloudant = require("@cloudant/cloudant");

const REDDIT_URL_PREFIX = "https://www.reddit.com"

function updateDestinationRedditThread(elements, params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const doc_id = params["id"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(params["destination_db"]);
    return db.get(doc_id).catch((err) => {
        console.log(err);
    }).then((doc) => {
        if (!doc) {
            doc = {}
        }
        doc._id = doc_id;
        doc.threads = elements;
        return db.insert(doc);
    });
}

function readSourceRedditThread(params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const element_id = params["id"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(params["dbname"]);
    return db.get(element_id).then((doc) => {
        return doc.thread;
    });
}

function reduceRedditThread(childThread) {
    let result = {};
    let data = childThread.data;

    result.title = data.title;
    result.text = data.selftext;
    result.url = `${REDDIT_URL_PREFIX}${data.permalink}`;
    result.upvotes = data.ups;
    result.downs = data.downs;
    result.link_flair = data.link_flair_text;
    return result;
}

function reduceRedditThreads(thread) {
    let result = thread.data.children.map((child) => {
        return reduceRedditThread(child);
    });
    return result;
}

function main(params) {
    return readSourceRedditThread(params)
        .then((thread) => {
            return reduceRedditThreads(thread);
        }).then((reducedThread) => {
            return updateDestinationRedditThread(reducedThread, params).then((data) => {
                return data;
            });
        })
}