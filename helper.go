package main

import (
	"sort"
	"time"
)

func SortKeys(mp map[string]StatusEntry, cb func(string, StatusEntry)) {
	keys := make([]string, 0, len(mp))

	for k := range mp {
		keys = append(keys, k)
	}

	// Down services first, then ordered by name
	sort.Slice(keys, func(i, j int) bool {
		a := keys[i]
		b := keys[j]

		aDown := mp[a].Status != 0
		bDown := mp[b].Status != 0

		if aDown && !bDown {
			return true
		} else if !aDown && bDown {
			return false
		}

		return a < b
	})

	for _, k := range keys {
		cb(k, mp[k])
	}
}

func _error(err error, t int64) StatusEntry {
	return StatusEntry{
		Status: time.Now().Unix(),
		Type:   "http",
		Error:  err.Error(),
		Time:   t,
	}
}

func _time(start time.Time) int64 {
	return time.Now().Sub(start).Milliseconds()
}
