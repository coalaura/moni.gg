package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"os"
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

const (
	DefaultFavicon     = "favicon.ico"
	DefaultBanner      = "banner.png"
	DefaultURL         = "https://example.com"
	DefaultTitle       = "Moni.GG Status-Page"
	DefaultDescription = "Stay informed with our status page. Tailored updates and real-time insights for a smooth experience. Keep a whisker's length ahead of any issues."
)

type EmbedFile struct {
	Name     string
	MimeType string
	Content  []byte

	Variables bool
}

var (
	//go:embed embed/*
	embedFS embed.FS

	files = []EmbedFile{
		{Name: "index.html", MimeType: "text/html", Variables: true},
		{Name: "main.css", MimeType: "text/css"},
		{Name: "main.js", MimeType: "application/javascript"},
	}
)

func ReBuildFrontend(cfg *Config) error {
	m := minify.New()

	htmlMin := html.Minifier{
		KeepDocumentTags: true,
		KeepQuotes:       true,
	}

	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", htmlMin.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	for i, file := range files {
		content, err := embedFS.ReadFile("embed/" + file.Name)
		if err != nil {
			return err
		}

		/*
			content, err = m.Bytes(file.MimeType, content)
			if err != nil {
				return err
			}
		*/

		file.Content = content

		files[i] = file
	}

	hash := _hash()

	for _, file := range files {
		if file.Variables {
			file.Content = _setVariables(file.Content, cfg, hash)
		}

		err := os.WriteFile("public/"+file.Name, file.Content, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func _setVariables(content []byte, cfg *Config, hash string) []byte {
	url := _default(cfg.TemplateURL, DefaultURL)
	banner := _default(cfg.TemplateBanner, DefaultBanner)

	content = bytes.ReplaceAll(content, []byte("{{banner}}"), _join(url, banner))
	content = bytes.ReplaceAll(content, []byte("{{url}}"), url)

	content = bytes.ReplaceAll(content, []byte("{{favicon}}"), _default(cfg.TemplateFavicon, DefaultFavicon))
	content = bytes.ReplaceAll(content, []byte("{{title}}"), _default(cfg.TemplateTitle, DefaultTitle))
	content = bytes.ReplaceAll(content, []byte("{{description}}"), _default(cfg.TemplateDescription, DefaultDescription))

	content = bytes.ReplaceAll(content, []byte("{{hash}}"), []byte(hash))

	return content
}

func _default(value string, def string) []byte {
	if value == "" {
		return []byte(def)
	}

	return []byte(value)
}

func _hash() string {
	hash := md5.New()

	for _, file := range files {
		hash.Write(file.Content)
	}

	return hex.EncodeToString(hash.Sum(nil))[0:8]
}

func _join(url []byte, path []byte) []byte {
	if bytes.HasPrefix(path, []byte("http")) {
		return path
	}

	if !bytes.HasSuffix(url, []byte("/")) {
		url = append(url, '/')
	}

	return append(url, path...)
}
