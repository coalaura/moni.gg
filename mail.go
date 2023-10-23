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

	email, title := BuildMail(entries, cfg.StatusPage)

	message.SetHeader("Subject", title)
	message.SetBody("text/html", email)

	message.Embed("public/banner.png")

	if strings.Contains(email, "cid:mail_up.png") {
		message.Embed("public/mail_up.png")
	}

	if strings.Contains(email, "cid:mail_down.png") {
		message.Embed("public/mail_down.png")
	}

	err := dialer.DialAndSend(message)

	if err == nil {
		log.Debug("Sent mail successfully")
	} else {
		log.Warning("Failed to send mail")
		log.WarningE(err)
	}
}

func BuildMail(entries map[string]StatusEntry, url string) (string, string) {
	var (
		down  int
		up    int
		index int

		title string
		body  string
	)

	SortKeys(entries, func(name string, entry StatusEntry) {
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

		if entry.Status == 0 {
			src = strings.ReplaceAll(src, "{{background}}", "#ebffeb")
			src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service is back online after %dms.", entry.Time))
			src = strings.ReplaceAll(src, "{{image}}", "cid:mail_up.png")

			up++
		} else {
			src = strings.ReplaceAll(src, "{{background}}", "#ffebeb")
			src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service went down after %dms with the error: <i>%s</i>.", entry.Time, entry.Error))
			src = strings.ReplaceAll(src, "{{image}}", "cid:mail_down.png")

			down++
		}

		body += src

		index++
	})

	if down > 0 && up > 0 {
		title = fmt.Sprintf("Status Alert (%d down, %d up)", down, up)
	} else if down > 0 {
		title = fmt.Sprintf("Status Alert (%d down)", down)
	} else {
		title = fmt.Sprintf("Status Alert (%d up)", up)
	}

	html := string(mainTemplate)

	html = strings.ReplaceAll(html, "{{url}}", url)
	html = strings.ReplaceAll(html, "{{banner}}", "cid:banner.png")

	html = strings.ReplaceAll(html, "{{body}}", body)

	return html, title
}
