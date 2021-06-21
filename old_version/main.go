package main

import (
	"os"
)

func main() {
	input := map[string]interface{}{
		"email_address": os.Getenv("EMAIL_ADDRESS"),
		"email_password": os.Getenv("EMAIL_PASSWORD"),
		"email_smtp_host": os.Getenv("EMAIL_SMTP_HOST"),
		"email_smtp_port": os.Getenv("EMAIL_SMTP_PORT"),
		"reddit_client_id": os.Getenv("REDDIT_CLIENT_ID"),
		"reddit_secret_token": os.Getenv("REDDIT_SECRET_TOKEN"),
		"reddit_username": os.Getenv("REDDIT_USERNAME"),
		"reddit_password": os.Getenv("REDDIT_PASSWORD"),
		"email_addresses": []interface{}{"test@example.com"},
	}
	Main(input)
}
