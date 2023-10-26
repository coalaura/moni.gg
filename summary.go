package main

import (
	"encoding/json"
	"os"
)

func UpdateSummary(status *StatusData) {
	total := len(status.Data)
	down := int(status.Down)

	b, _ := json.Marshal(map[string]int{
		"total":   total,
		"online":  total - down,
		"offline": down,
	})

	_ = os.WriteFile("public/summary.json", b, 0777)
}
