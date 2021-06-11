# wsbstockapp

NOTE: the code was written this way since I'm running the app on IBM Cloud, as can be seen from the [app.go Main method](./app.go#L304-L311).

### Getting the necessary credentials for Reddit:
Follow the directions at the [Reddit OAuth2 wiki](https://github.com/reddit-archive/reddit/wiki/OAuth2) to get the necessary credentials for the Reddit API.

### Running the application
Go into the [main.go](./main.go) file and update the `email_addresses` you would want to send emails to.

You can choose to run it as is after setting the following environment variables:
- `EMAIL_ADDRESS`
- `EMAIL_PASSWORD`
- `EMAIL_SMTP_HOST`
- `EMAIL_SMTP_PORT`
- `REDDIT_CLIENT_ID`
- `REDDIT_SECRET_TOKEN`
- `REDDIT_USERNAME`
- `REDDIT_PASSWORD`