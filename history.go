package main

import "time"

func (h *History) TrackHistoric(isUp bool) {
	h.Cleanup()

	now := time.Now().UTC()
	since := h.Since(now)

	h.CheckedAt = now.Unix()

	if isUp {
		return
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Unix()

	for h.FloorTime(since).Unix() < today {
		next := h.FloorTime(since).AddDate(0, 0, 1)

		h.Add(since, next)

		since = next
	}

	h.Add(since, now)
}

func (h *History) Add(since, now time.Time) {
	key := since.UTC().Format("2006-01-02")

	minutes := now.Sub(since).Minutes()

	h.Downtimes[key] += int64(minutes)
}

func (h *History) FloorTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (h *History) Since(now time.Time) time.Time {
	if h.CheckedAt == 0 {
		return now.Add(-1 * time.Minute)
	}

	return time.Unix(h.CheckedAt, 0)
}

func (h *History) Cleanup() {
	if h.Downtimes == nil {
		h.Downtimes = make(map[string]int64)
	}

	unixMin := time.Now().AddDate(0, 0, -90).Unix()

	for day := range h.Downtimes {
		d, err := time.Parse("2006-01-02", day)

		if err != nil || d.Unix() < unixMin {
			delete(h.Downtimes, day)
		}
	}
}
