package main

import (
	"encoding/json"
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

	log.Info("Re-building frontend...")
	err = ReBuildFrontend(mainConfig)
	log.MustPanic(err)

	log.Info("Loading tasks...")
	tasks, err := LoadTasks()
	log.MustPanic(err)

	log.Info("Reading previous status...")
	status, err := ReadPrevious(tasks)
	log.MustPanic(err)

	// Test mail sending
	if len(os.Args) > 1 && os.Args[1] == "mail" {
		SendExampleMail(mainConfig)

		return
	}

	status.Down = 0

	var (
		mutex sync.Mutex
		wg    sync.WaitGroup
	)

	for name, task := range tasks {
		log.DebugF("Checking %s...\n", name)

		previous, ok := status.Data[name]
		previousStatus := ok && previous.Error == ""

		wg.Add(1)

		go func(name string, task Task) {
			err := task.Resolve()

			currentStatus := err.Error == ""

			if ok {
				err.History = previous.History
			}

			err.History.TrackHistoric(currentStatus)

			if currentStatus != previousStatus {
				err._new = true
			}

			if !currentStatus {
				log.Warning(err.Error)

				status.Down++
			}

			mutex.Lock()
			status.Data[name] = err
			mutex.Unlock()

			wg.Done()
		}(name, task)
	}

	wg.Wait()

	if status.ShouldSendMail() {
		SendMail(status, mainConfig)
	}

	status.Time = time.Now().Unix()

	log.Info("Saving status data...")
	jsn, err := json.Marshal(status)
	log.MustPanic(err)

	_ = os.WriteFile("status.json", jsn, 0777)
	_ = os.WriteFile("public/status.json", jsn, 0777)
}
