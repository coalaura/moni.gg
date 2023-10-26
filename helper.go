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

		aDown := !mp[a].Operational
		bDown := !mp[b].Operational

		aNew := mp[a]._new
		bNew := mp[b]._new

		if aDown && !bDown {
			return true
		} else if !aDown && bDown {
			return false
		}

		if aNew && !bNew {
			return true
		} else if !aNew && bNew {
			return false
		}

		return a < b
	})

	for _, k := range keys {
		cb(k, mp[k])
	}
}

func _error(err error, responseTime int64) StatusEntry {
	return StatusEntry{
		Operational:  false,
		Type:         "http",
		Error:        err.Error(),
		ResponseTime: responseTime,
	}
}

func _time(start time.Time) int64 {
	return time.Now().Sub(start).Milliseconds()
}
