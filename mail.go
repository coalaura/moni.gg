package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

var (
	//go:embed email/main.html
	mainTemplate []byte

	//go:embed email/service_left.html
	serviceTemplateLeft []byte

	//go:embed email/service_right.html
	serviceTemplateRight []byte

	dialer *gomail.Dialer

	colors = []string{
		"rgba(255, 153, 153, 0.2)",
		"rgba(255, 221, 153, 0.2)",
		"rgba(221, 255, 153, 0.2)",
		"rgba(153, 255, 153, 0.2)",
		"rgba(153, 255, 221, 0.2)",
		"rgba(153, 221, 255, 0.2)",
		"rgba(153, 153, 255, 0.2)",
		"rgba(221, 153, 255, 0.2)",
		"rgba(255, 153, 221, 0.2)",
	}
)

func SendMail(entries map[string]StatusEntry, cfg *Config) {
	if cfg.EmailTo == "" || cfg.SMTPHost == "" || cfg.SMTPUser == "" || cfg.SMTPPassword == "" || cfg.SMTPPort == 0 {
		return
	}

	if dialer == nil {
		dialer = gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	message := gomail.NewMessage()

	message.SetHeader("From", cfg.SMTPUser)
	message.SetHeader("To", cfg.EmailTo)

	email, title := BuildMail(entries)

	message.SetHeader("Subject", title)
	message.SetBody("text/html", email)

	message.Embed("public/banner.png")

	if strings.Contains(email, "cid:email_up.png") {
		message.Embed("public/email_up.png")
	}

	if strings.Contains(email, "cid:email_down.png") {
		message.Embed("public/email_down.png")
	}

	err := dialer.DialAndSend(message)

	if err == nil {
		log.Debug("Sent mail successfully")
	} else {
		log.Warning("Failed to send mail")
		log.WarningE(err)
	}
}

func BuildMail(entries map[string]StatusEntry) (string, string) {
	var (
		down  int
		up    int
		index int

		title string
		body  string
	)

	for name, entry := range entries {
		var (
			src string
		)

		if index%2 == 0 {
			src = string(serviceTemplateRight)
		} else {
			src = string(serviceTemplateLeft)
		}

		src = strings.ReplaceAll(src, "{{name}}", name)
		src = strings.ReplaceAll(src, "{{type}}", entry.Type)
		src = strings.ReplaceAll(src, "{{background}}", colors[index%len(colors)])

		if entry.Status == 0 {
			src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service is back online after %dms.", entry.Time))
			src = strings.ReplaceAll(src, "{{image}}", "cid:email_up.png")
		} else {
			src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service went down after %dms with the error: <i>%s</i>.", entry.Time, entry.Error))
			src = strings.ReplaceAll(src, "{{image}}", "cid:email_down.png")
		}

		body += src

		index++
	}

	if down > 0 && up > 0 {
		title = fmt.Sprintf("Status Alert (%d down, %d up)", down, up)
	} else if down > 0 {
		title = fmt.Sprintf("Status Alert (%d down)", down)
	} else if up > 0 {
		title = fmt.Sprintf("Status Alert (%d up)", up)
	} else {
		title = "Status Alert"
	}

	html := string(mainTemplate)

	html = strings.ReplaceAll(html, "{{title}}", title)
	html = strings.ReplaceAll(html, "{{banner}}", "cid:banner.png")

	html = strings.ReplaceAll(html, "{{body}}", body)

	return html, title
}
