const Openwhisk = require('openwhisk');

const WHISK_ALARM_FEED_PATH = "/whisk.system/alarms/alarm"
const CRON_FORMAT = "%d %d * * *";

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

function setupSubredditTriggersAndRules(subreddits, whisk) {
    let startTime = Date.UTC(2021, 7, 16, 12, 0, 0);
    const namespace = "subreddit";
    const actionName = "get-subreddit-top-threads";
    let triggers = subreddits.map((subreddit) => {
        const feedParams = {cron: CRON_FORMAT.format(startTime.getUTCMinutes(), startTime.getUTCHours()), trigger_payload: {subreddit}};
        const triggerName = "subreddit-caller-trigger-%d-%d".format(startTime.getUTCHours(), startTime.getUTCMinutes());
        const ruleName = "subreddit-caller-rule-%d-%d".format(startTime.getUTCHours(), startTime.getUTCMinutes());
        const feedName = "subreddit-caller-feed-%d-%d".format(startTime.getUTCHours(), startTime.getUTCMinutes());
        const nextMinute = startTime.getUTCMinutes() + 1;
        const nextHour = startTime.getUTCHours + Math.floor(nextMinute/60);
        startTime.setUTCMinutes(nextMinute % 60);
        startTime.setUTCHours(nextHour);
        const payload = {subreddit};
        const annotations = {
            wsbOrigin: true
        };
        return whisk.triggers.update({
            name: triggerName,
            namespace: namespace,
            trigger: payload,
            annotations: annotations
        }).then((trigger) => {
            return Promise.all([
                whisk.feeds.update({
                    feedName: WHISK_ALARM_FEED_PATH,
                    trigger: triggerName,
                    namespace: namespace,
                    params: feedParams
                }),
                whisk.rules.update({
                    name: ruleName,
                    action: actionName,
                    trigger: triggerName,
                    namespace: namespace,
                    annotations: annotations
                })
            ]);
        });
    });
    return Promise.all(triggers);
}

function main(params) {
    if (params["_id"] !== "recipients") {
        return;
    }

    const subreddits = getListOfSubreddits(params["users"]);
    const whisk = Openwhisk();
    return setupSubredditTriggersAndRules(subreddits, whisk);
}