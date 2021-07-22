const Openwhisk = require('openwhisk');

const WHISK_ALARM_FEED_PATH = "/whisk.system/alarms/alarm";

function getListOfSubreddits(users) {
    let resultMap = {};
    users.forEach((user) => {
        const subreddits = user["subreddits"];
        subreddits.forEach((data) => {
            resultMap[data.subreddit.toLowerCase()] = true;
        })
    })
    return Object.keys(resultMap);
}

function createOrUpdate(whiskResource, options) {
    return whiskResource.create(options)
        .catch((error) => {
            console.log(error);
            return whiskResource.update(options);
        })
}

function createTrigger(actionName, startTime, payload, whisk) {
    const cronTab = `${startTime.getUTCMinutes()} ${startTime.getUTCHours()} * * *`;
    const feedParams = {cron:cronTab, trigger_payload: payload};
    const triggerName = `subreddit-caller-trigger-${startTime.getUTCHours()}-${startTime.getUTCMinutes()}`;
    const ruleName = `subreddit-caller-rule-${startTime.getUTCHours()}-${startTime.getUTCMinutes()}`;
    const annotations = {
        wsbOrigin: true
    };
    return createOrUpdate(whisk.triggers,{
        triggerName: triggerName,
        annotations: annotations
    }).then((trigger) => {
        return Promise.all([
            createOrUpdate(whisk.feeds,{
                feedName: WHISK_ALARM_FEED_PATH,
                trigger: triggerName,
                params: feedParams
            }),
            createOrUpdate(whisk.rules, {
                ruleName: ruleName,
                action: actionName,
                trigger: triggerName,
                annotations: annotations
            })
        ]);
    });
}

function setupSubredditTriggersAndRules(subreddits, whisk) {
    let startTime = new Date(Date.UTC(2021, 7, 16, 12, 0, 0));
    const package = "reddit";
    const actionName = `${package}/get-subreddit-top-threads`;
    const newsletterActionName = `${package}/get-users-and-send-newsletter-sequence`;
    let triggers = subreddits.map((subreddit) => {
        const payload = {subreddit: subreddit};
        const trigger = createTrigger(actionName, startTime, payload, whisk);
        const nextMinute = startTime.getUTCMinutes() + 1;
        const nextHour = startTime.getUTCHours() + Math.floor(nextMinute/60);
        startTime.setUTCMinutes(nextMinute % 60);
        startTime.setUTCHours(nextHour);
        return trigger;
    });
    return Promise.all(triggers).then((results) => {
        return createTrigger(newsletterActionName, startTime, {
            dbname: "reddit-users",
            id: "recipients"
        }, whisk);
    });
}

function main(params) {
    if (params["_id"] !== "recipients") {
        return;
    }

    const subreddits = getListOfSubreddits(params["users"]);
    const whisk = Openwhisk();
    return setupSubredditTriggersAndRules(subreddits, whisk).then((results) => {
        return {success: true}
    });
}