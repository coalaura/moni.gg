package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
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

var (
	//go:embed embed/*
	embedFS embed.FS
)

func ReBuildFrontend(cfg *Config) error {
	files, err := embedFS.ReadDir("embed")
	if err != nil {
		return err
	}

	h := _hash(files)

	m := minify.New()

	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

	var (
		html bytes.Buffer
		css  bytes.Buffer
		js   bytes.Buffer
	)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		mime := _mime(name)

		content, err := embedFS.ReadFile("embed/" + name)
		if err != nil {
			return err
		}

		content = _setVariables(content, cfg, h)

		content, err = m.Bytes(mime, content)
		if err != nil {
			return err
		}

		switch mime {
		case "text/html":
			html.Write(content)
		case "text/css":
			css.Write(content)
			css.Write([]byte("\n"))
		case "application/javascript":
			js.Write(content)
			js.Write([]byte("\n"))
		}
	}

	err = os.WriteFile("public/index.html", html.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile("public/main.css", css.Bytes(), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile("public/main.js", js.Bytes(), 0644)
	if err != nil {
		return err
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

func _hash(files []fs.DirEntry) string {
	hash := md5.New()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, _ := embedFS.ReadFile("embed/" + file.Name())

		hash.Write(content)
	}

	return hex.EncodeToString(hash.Sum(nil))[0:8]
}

func _mime(name string) string {
	switch filepath.Ext(name) {
	case ".css":
		return "text/css"
	case ".html":
		return "text/html"
	case ".js":
		return "application/javascript"
	}

	return ""
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
