const Openwhisk = require('openwhisk');

function getListOfSubreddits(users) {
    let resultMap = {};
    users.forEach((user) => {
        const subreddits = user["subreddits"];
        subreddits.forEach((subreddit) => {
            resultMap[subreddit.toLowerCase()] = true;
        })
    })
    return Object.keys(resultMap);
}

function logPromise(prefix, promise) {
    return promise.then((result) => {
        console.log(`${prefix}: ${JSON.stringify(result)}`);
    })
}

function main(params) {
    /*if (params["_id"] !== "recipients") {
        return;
    }

    const subreddits = getListOfSubreddits(params["users"]);*/
    const whisk = Openwhisk();
    return Promise.all([
        logPromise("Triggers", whisk.triggers.list()),
        logPromise("Actions", whisk.actions.list()),
        logPromise("Rules", whisk.rules.list())
        ]
    )
}