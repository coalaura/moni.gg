package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	StatusPage string

	EmailTo string

	SMTPHost     string
	SMTPPort     int
	SMTPPassword string
	SMTPUser     string
}

type StatusEntry struct {
	Status   int64         `json:"status"`
	Type     string        `json:"type"`
	Error    string        `json:"error,omitempty"`
	Historic map[int64]int `json:"historic,omitempty"`
	Time     int64         `json:"time"`

	New bool
}

type StatusJSON struct {
	Time int64                  `json:"time"`
	Data map[string]StatusEntry `json:"data"`
	Down int64                  `json:"down"`

	New int
}

type SmallJSON struct {
	Total   int64 `json:"total"`
	Online  int64 `json:"online"`
	Offline int64 `json:"offline"`
}

func ReadPrevious(tasks map[string]Task) (*StatusJSON, error) {
	_, err := os.Stat("status.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &StatusJSON{
				Time: time.Now().Unix(),
				Data: make(map[string]StatusEntry),
			}, nil
		}

		return nil, err
	}

	b, _ := os.ReadFile("status.json")

	var status StatusJSON
	err = json.Unmarshal(b, &status)
	if err != nil {
		return nil, err
	}

	// Cleanup old data
	for name := range status.Data {
		if _, ok := tasks[name]; !ok {
			delete(status.Data, name)
		}
	}

	return &status, nil
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
	}, nil
}

func ReadConfigs() (map[string]Task, error) {
	configs := make(map[string]Task)

	err := filepath.Walk("./config", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)

		if ext != ".http" && ext != ".mysql" {
			return nil
		}

		name := strings.Split(filepath.Base(path), ".")[0]

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := strings.ReplaceAll(string(data), "\r\n", "\n")

		switch ext {
		case ".http":
			configs[name] = NewHTTPTask(content)
		case ".mysql":
			configs[name] = NewMySQLTask(content)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return configs, nil
}
