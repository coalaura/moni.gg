package main

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
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

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		content, err := embedFS.ReadFile("embed/" + file.Name())
		if err != nil {
			return err
		}

		content = _setVariables(content, cfg, h)

		err = os.WriteFile("public/"+file.Name(), content, 0777)
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

func _join(url []byte, path []byte) []byte {
	if bytes.HasPrefix(path, []byte("http")) {
		return path
	}

	if !bytes.HasSuffix(url, []byte("/")) {
		url = append(url, '/')
	}

	return append(url, path...)
}
