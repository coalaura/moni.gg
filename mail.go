package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

var (
	//go:embed email/main.html
	mainTemplate []byte

	//go:embed email/service_left.html
	serviceTemplateLeft []byte

	//go:embed email/service_right.html
	serviceTemplateRight []byte

	//go:embed email/info.html
	infoTemplate []byte

	//go:embed email/error.html
	errorTemplate []byte

	dialer *gomail.Dialer
)

func SendMail(data *StatusJSON, cfg *Config) {
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

	email := BuildMail(data.Data, cfg.StatusPage)

	message.SetHeader("Subject", fmt.Sprintf("Status Alert (%d down, %d up)", data.Down, len(data.Data)-int(data.Down)))
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

func BuildMail(entries map[string]StatusEntry, url string) string {
	var (
		index int
		body  string
	)

	SortKeys(entries, func(name string, entry StatusEntry) {
		var src string

		if entry.New {
			if index%2 == 0 {
				src = string(serviceTemplateRight)
			} else {
				src = string(serviceTemplateLeft)
			}

			src = strings.ReplaceAll(src, "{{type}}", strings.ToLower(entry.Type))

			if entry.Status == 0 {
				src = strings.ReplaceAll(src, "{{background}}", "#d6ffd6")
				src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service is back online after <b>%dms</b>.", entry.Time))
				src = strings.ReplaceAll(src, "{{image}}", "cid:mail_up.png")
			} else {
				err := string(errorTemplate)
				err = strings.ReplaceAll(err, "{{error}}", entry.Error)

				src = strings.ReplaceAll(src, "{{background}}", "#ffd6d6")
				src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service went down after <b>%dms</b>. %s", entry.Time, err))
				src = strings.ReplaceAll(src, "{{image}}", "cid:mail_down.png")
			}

			index++
		} else {
			src = string(infoTemplate)

			src = strings.ReplaceAll(src, "{{name}}", name)

			if entry.Status == 0 {
				src = strings.ReplaceAll(src, "{{background}}", "#b3ffb3")
				src = strings.ReplaceAll(src, "{{text}}", "Online")
			} else {
				src = strings.ReplaceAll(src, "{{background}}", "#ffb3b3")
				src = strings.ReplaceAll(src, "{{text}}", "Still Offline")
			}
		}

		src = strings.ReplaceAll(src, "{{name}}", name)

		body += src
	})

	html := string(mainTemplate)

	html = strings.ReplaceAll(html, "{{url}}", url)
	html = strings.ReplaceAll(html, "{{banner}}", "cid:banner.png")
	html = strings.ReplaceAll(html, "{{time}}", time.Now().Format("1/2/2006 - 3:04:05 PM MST"))

	html = strings.ReplaceAll(html, "{{body}}", body)

	return html
}

func SendExampleMail(cfg *Config) {
	entries := map[string]StatusEntry{
		"Alpha": {
			Type:   "HTTP",
			Status: 0,
			Error:  "",
			Time:   int64(rand.Intn(1000)),
			New:    true,
		},
		"Charlie": {
			Type:   "HTTP",
			Status: 0,
			Error:  "",
			Time:   int64(rand.Intn(1000)),
		},
		"Bravo": {
			Type:   "HTTP",
			Status: time.Now().Unix() - int64(rand.Intn(60*60)),
			Error:  "Failed to connect to host",
			Time:   int64(rand.Intn(1000)),
			New:    true,
		},
		"Delta": {
			Type:   "HTTP",
			Status: time.Now().Unix() - int64(rand.Intn(60*60)),
			Error:  "Failed to connect to host",
			Time:   int64(rand.Intn(1000)),
		},
	}

	SendMail(&StatusJSON{
		Data: entries,
		Down: 2,
	}, cfg)
}
