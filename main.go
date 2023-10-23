package main

import (
	"encoding/json"
	"math"
	"os"
	"sync"
	"time"

	logger_v2 "gitlab.com/milan44/logger-v2"
)

var (
	log = logger_v2.NewColored()
)

func main() {
	CheckExistence()

	_ = os.MkdirAll("public", 0777)
	_ = os.MkdirAll("config", 0777)

	log.Info("Loading config...")
	mainConfig, err := ReadMainConfig()
	log.MustPanic(err)

	log.Info("Reading .envs...")
	tasks, err := ReadConfigs()
	log.MustPanic(err)

	log.Info("Reading previous status...")
	status, err := ReadPrevious()
	log.MustPanic(err)

	// Test mail sending
	if len(os.Args) > 1 && os.Args[1] == "mail" {
		SendExampleMail(mainConfig)

		return
	}

	status.Down = 0

	minute := getMinute()

	oldestMinute := minute - (144 * 5)

	var (
		mutex sync.Mutex
		wg    sync.WaitGroup

		emails = make(map[string]StatusEntry)
	)

	data := SmallJSON{}

	for name, task := range tasks {
		log.DebugF("Checking %s...\n", name)

		data.Total++

		previous, ok := status.Data[name]
		previousStatus := ok && previous.Error == ""

		wg.Add(1)

		go func(name string, task Task) {
			err := task.Resolve()

			newStatus := err.Error == ""

			if ok {
				err.Historic = previous.Historic
			}

			if err.Historic == nil {
				err.Historic = make(map[int64]int)
			}

			if newStatus != previousStatus {
				emails[name] = err
			}

			if !newStatus {
				log.Warning(err.Error)

				data.Offline++

				status.Down++

				err.Historic[minute]++
			} else {
				data.Online++
			}

			if ok && previous.Status > 0 && err.Status > 0 {
				err.Status = previous.Status
			}

			for min := range err.Historic {
				if min < oldestMinute {
					delete(err.Historic, min)
				}
			}

			mutex.Lock()
			status.Data[name] = err
			mutex.Unlock()

			wg.Done()
		}(name, task)
	}

	wg.Wait()

	if len(emails) > 0 {
		SendMail(emails, mainConfig)
	}

	// Cleanup old data
	for name := range status.Data {
		if _, ok := tasks[name]; !ok {
			delete(status.Data, name)
		}
	}

	status.Time = time.Now().Unix()

	log.Info("Saving status data...")
	jsn, err := json.Marshal(status)
	log.MustPanic(err)

	shrt, err := json.Marshal(data)
	log.MustPanic(err)

	_ = os.WriteFile("status.json", jsn, 0777)

	_ = os.WriteFile("public/status.json", jsn, 0777)
	_ = os.WriteFile("public/summary.json", shrt, 0777)
}

func getMinute() int64 {
	return int64(math.Floor(float64(time.Now().Unix()) / 600.0))
}
