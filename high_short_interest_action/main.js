const HIGH_SHORT_INTEREST_ENDPOINT = "https://www.highshortinterest.com/all/";
const TABLE_ROW_PATTERN = /<tr>\s*<td[^>]*><a\s.*?href="(.*?)"[^>]*?>([A-Z]+)<\/a><\/td>\s*<td[^>]*>(.+?)<\/td>\s*<td[^>]*>(.+?)<\/td>\s*<td[^>]*>([0-9.]+?)%<\/td>\s*<td[^>]*>([0-9A-Za-z.]+?)<\/td>\s*<td[^>]*>([0-9A-Za-z.]+?)<\/td>\s*<td[^>]*>(.+?)<\/td>\s*<\/tr>/g;

const Cloudant = require("@cloudant/cloudant");
const axios = require("axios").default;

function getHighShortInterestStocks(params) {
    return axios.get(HIGH_SHORT_INTEREST_ENDPOINT).then((html_response) => {
        let result = [];
        const data = html_response.data;
        let match;
        do {
            match = TABLE_ROW_PATTERN.exec(data);
            if (match) {
                result.push({
                    url: match[1],
                    symbol: match[2],
                    name: match[3],
                    exchange: match[4],
                    interest: match[5],
                    floating: match[6],
                    outstanding: match[7],
                    industry: match[8]
                });
            }
        } while(match)
        return result;
    })
}

function uploadHighShortInterestStocks(data, params) {
    const cloudant_url = params["cloudant_url"];
    const cloudant_apikey = params["cloudant_apikey"];
    const document_id = params["document_id"]
    const cloudant = Cloudant({url: cloudant_url, maxAttempt: 5, plugins: [{ iamauth: { iamApiKey: cloudant_apikey } }, { retry: { retryDelayMultiplier: 4 } }]});
    const db = cloudant.use(params["dbname"]);
    return db.get(document_id).catch((error) => {
        console.log(error);
    }).then((doc) => {
        if (!doc) {
            doc = {}
        }
        doc._id = document_id;
        doc.data = data;
        return db.insert(doc);
    });
}

function main(params) {
    return getHighShortInterestStocks(params).then((data) => {
        return uploadHighShortInterestStocks(data, params);
    });
}