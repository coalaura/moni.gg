package main

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLTask struct {
	Hostname string
	Port     string
	Username string
	Password string
	Database string

	Query string
}

func NewMySQLTask(content string) *MySQLTask {
	task := &MySQLTask{}

	lines := strings.Split(content, "\n")

	rgx := regexp.MustCompile(`\s*;\s*`)
	config := rgx.Split(lines[0], -1)

	for _, entry := range config {
		pair := strings.Split(entry, "=")

		switch strings.ToLower(pair[0]) {
		case "hostname":
			task.Hostname = pair[1]
		case "port":
			task.Port = pair[1]
		case "username":
			task.Username = pair[1]
		case "password":
			task.Password = pair[1]
		case "database":
			task.Database = pair[1]
		}
	}

	task.Query = strings.TrimSpace(strings.Join(lines[1:], "\n"))

	if task.Query == "" {
		task.Query = "SELECT 1"
	}

	if task.Port == "" {
		task.Port = "3306"
	}

	return task
}

func (m *MySQLTask) Resolve() StatusEntry {
	resp := _mysql(m.Hostname, m.Port, m.Username, m.Password, m.Database, m.Query)

	if resp.Error != "" {
		time.Sleep(10 * time.Second)

		resp = _mysql(m.Hostname, m.Port, m.Username, m.Password, m.Database, m.Query)
	}

	return resp
}

func _mysql(hostname, port, username, password, database, query string) StatusEntry {
	start := time.Now()

	db, err := sql.Open("mysql", username+":"+password+"@tcp("+hostname+":"+port+")/"+database)
	if err != nil {
		return _error(err, _time(start))
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		return _error(err, _time(start))
	}

	rows, err := db.Query(query)
	if err != nil {
		return _error(err, _time(start))
	}

	defer rows.Close()

	if !rows.Next() {
		return _error(errors.New("No rows returned"), _time(start))
	}

	return StatusEntry{
		Operational:  true,
		Type:         "mysql",
		ResponseTime: _time(start),
	}
}
