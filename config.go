package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	StatusPage string

	EmailTo string

	SMTPHost     string
	SMTPPort     int
	SMTPPassword string
	SMTPUser     string

	TemplateFavicon     string
	TemplateBanner      string
	TemplateURL         string
	TemplateTitle       string
	TemplateDescription string
}

func ReadMainConfig() (*Config, error) {
	data, err := os.ReadFile(".env")
	if err != nil {
		return nil, err
	}

	env, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(env["SMTP_PORT"])

	return &Config{
		StatusPage: env["STATUS_PAGE"],

		EmailTo:      env["EMAIL_TO"],
		SMTPHost:     env["SMTP_HOST"],
		SMTPPort:     port,
		SMTPUser:     env["SMTP_USER"],
		SMTPPassword: env["SMTP_PASSWORD"],

		TemplateFavicon:     env["TEMPLATE_FAVICON"],
		TemplateBanner:      env["TEMPLATE_BANNER"],
		TemplateURL:         env["TEMPLATE_URL"],
		TemplateTitle:       env["TEMPLATE_TITLE"],
		TemplateDescription: env["TEMPLATE_DESCRIPTION"],
	}, nil
}
