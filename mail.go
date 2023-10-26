package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
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

	//go:embed email/banner.png
	bannerImage []byte

	//go:embed email/mail_up.png
	mailUpImage []byte

	//go:embed email/mail_down.png
	mailDownImage []byte

	dialer *gomail.Dialer
)

func SendMail(data *StatusData, cfg *Config) {
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

	EmbedImage(message, "banner.png", bannerImage)

	if strings.Contains(email, "cid:mail_up.png") {
		EmbedImage(message, "mail_up.png", mailUpImage)
	}

	if strings.Contains(email, "cid:mail_down.png") {
		EmbedImage(message, "mail_down.png", mailDownImage)
	}

	err := dialer.DialAndSend(message)

	if err == nil {
		log.Debug("Sent mail successfully")
	} else {
		log.Warning("Failed to send mail")
		log.WarningE(err)
	}
}

func EmbedImage(message *gomail.Message, name string, data []byte) {
	message.Embed(name, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(data)

		return err
	}))
}

func BuildMail(entries map[string]StatusEntry, url string) string {
	var (
		index int
		body  string
	)

	SortKeys(entries, func(name string, entry StatusEntry) {
		var src string

		if entry._new {
			if index%2 == 0 {
				src = string(serviceTemplateRight)
			} else {
				src = string(serviceTemplateLeft)
			}

			src = strings.ReplaceAll(src, "{{type}}", strings.ToLower(entry.Type))

			if entry.Operational {
				src = strings.ReplaceAll(src, "{{background}}", "#d6fff2")
				src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service is back operational after <b>%dms</b>.", entry.ResponseTime))
				src = strings.ReplaceAll(src, "{{image}}", "cid:mail_up.png")
			} else {
				err := string(errorTemplate)
				err = strings.ReplaceAll(err, "{{error}}", entry.Error)

				src = strings.ReplaceAll(src, "{{background}}", "#ffdcd6")
				src = strings.ReplaceAll(src, "{{text}}", fmt.Sprintf("Service went down after <b>%dms</b>. %s", entry.ResponseTime, err))
				src = strings.ReplaceAll(src, "{{image}}", "cid:mail_down.png")
			}

			index++
		} else {
			src = string(infoTemplate)

			src = strings.ReplaceAll(src, "{{name}}", name)

			if entry.Operational {
				src = strings.ReplaceAll(src, "{{background}}", "#b3ffe7")
				src = strings.ReplaceAll(src, "{{text}}", "Operational")
			} else {
				src = strings.ReplaceAll(src, "{{background}}", "#ffbeb3")
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
			Type:         "HTTP",
			Operational:  true,
			Error:        "",
			ResponseTime: int64(rand.Intn(1000)),
			_new:         true,
		},
		"Charlie": {
			Type:         "HTTP",
			Operational:  true,
			Error:        "",
			ResponseTime: int64(rand.Intn(1000)),
		},
		"Bravo": {
			Type:         "HTTP",
			Error:        "Failed to connect to host",
			ResponseTime: int64(rand.Intn(1000)),
			_new:         true,
		},
		"Delta": {
			Type:         "HTTP",
			Error:        "Failed to connect to host",
			ResponseTime: int64(rand.Intn(1000)),
		},
	}

	SendMail(&StatusData{
		Data: entries,
		Down: 2,
	}, cfg)
}
