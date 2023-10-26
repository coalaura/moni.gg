package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Task interface {
	Resolve() StatusEntry
}

func LoadTasks() (map[string]Task, error) {
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
