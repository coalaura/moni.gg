package main

import (
	"encoding/json"
	"os"
	"time"
)

type StatusEntry struct {
	Operational bool   `json:"operational"`
	Type        string `json:"type"`

	Error        string `json:"error,omitempty"`
	ResponseTime int64  `json:"response_time"`

	History History `json:"history"`

	_new bool
}

type StatusData struct {
	Time int64                  `json:"time"`
	Data map[string]StatusEntry `json:"data"`
	Down int64                  `json:"down"`
}

func (s *StatusData) ShouldSendMail() bool {
	for _, entry := range s.Data {
		if entry._new {
			return true
		}
	}

	return false
}

func ReadPrevious(tasks map[string]Task) (*StatusData, error) {
	_, err := os.Stat("status.json")
	if err != nil {
		if os.IsNotExist(err) {
			return &StatusData{
				Time: time.Now().Unix(),
				Data: make(map[string]StatusEntry),
			}, nil
		}

		return nil, err
	}

	b, _ := os.ReadFile("status.json")

	var status StatusData
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
