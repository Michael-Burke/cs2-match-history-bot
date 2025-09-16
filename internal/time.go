package internal

import (
	"log"
	"os"
	"time"
)

func ToUnixMillis(t time.Time) int64 { return t.UTC().UnixMilli() }

func CurrentWeekWindow(now time.Time) (start, end int64, human_start, human_end string) {
	loadEnv(true)
	location := os.Getenv("TIME_ZONE")
	log.Println("Time zone: ", location)
	if location == "" {
		location = "US/Eastern"
	}
	loc, _ := time.LoadLocation(location)
	n := now.In(loc)

	// Find Monday of the current week (00:00:00)
	// Go: Monday = 1 ... Sunday = 0
	weekday := int(n.Weekday())
	if weekday == 0 { // Sunday -> treat as 7
		weekday = 7
	}
	// Back to Monday
	monday := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc).
		AddDate(0, 0, -(weekday - 1))

	human_start = monday.Format("01/02/2006")
	human_end = monday.AddDate(0, 0, 7).Format("01/02/2006")

	start = ToUnixMillis(monday)
	end = ToUnixMillis(monday.AddDate(0, 0, 7)) // next Monday 00:00

	return start, end, human_start, human_end
}
