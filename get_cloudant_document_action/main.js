const Cloudant = require("@cloudant/cloudant");

function main(params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const document_id = params["id"];
    const dbname = params["dbname"];
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(dbname);
    return db.get(document_id);
}