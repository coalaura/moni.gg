package main

import (
	"errors"
	"github.com/mitchellh/go-ps"
	"os"
	"path/filepath"
)

func CheckExistence() {
	list, err := ps.Processes()
	log.MustPanic(err)

	pid := os.Getpid()

	filename := filepath.Base(os.Args[0])

	for _, process := range list {
		if process.Pid() != pid && filename == process.Executable() {
			log.MustPanic(errors.New("already started"))
		}
	}
}
