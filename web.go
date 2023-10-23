package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Resolve(lines []string) StatusEntry {
	header := strings.Split(lines[0], " ")
	headers := make(map[string]string)

	data := make([]string, 0)

	headerOver := false

	for _, line := range lines[1:] {
		if headerOver {
			data = append(data, line)
		} else {
			if line == "" {
				headerOver = true
			} else {
				entry := strings.Split(line, ": ")

				headers[entry[0]] = entry[1]
			}
		}
	}

	resp := _request(header[0], "https://"+headers["Host"]+header[1], strings.Join(data, "\n"), headers)

	if resp.Error != "" {
		time.Sleep(10 * time.Second)

		return _request(header[0], "https://"+headers["Host"]+header[1], strings.Join(data, "\n"), headers)
	}

	return resp
}

func _request(method, url, data string, headers map[string]string) StatusEntry {
	body := strings.NewReader(data)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return _error(err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return _error(err)
	}

	if resp.StatusCode != 200 {
		return _error(errors.New(fmt.Sprintf("Status code was %d instead of 200", resp.StatusCode)))
	}

	return StatusEntry{
		Status: 0,
		Type:   "http",
		Error:  "",
	}
}

func _error(err error) StatusEntry {
	return StatusEntry{
		Status: time.Now().Unix(),
		Type:   "http",
		Error:  err.Error(),
	}
}
