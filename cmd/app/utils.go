package main

import (
	"fmt"
	"time"
)

func parseDueDate(input string) int64 {
	if input == "" {
		return 0
	}

	// Try parsing as number of days from now
	var days int
	if len(input) <= 3 {
		_, err := fmt.Sscanf(input, "%d", &days)
		if err == nil && days > 0 {
			t := time.Now().AddDate(0, 0, days)
			year, month, day := t.Date()
			t = time.Date(year, month, day, 23, 59, 59, 0, time.Local)
			return t.Unix()
		}
	}

	// Try parsing as date format MM/DD/YY
	t, err := time.Parse("1/2/2006", input)
	if err == nil {
		year, month, day := t.Date()
		t = time.Date(year, month, day, 23, 59, 59, 0, time.Local)
		return t.Unix()
	}
	// Try parsing as date format MM/DD/YY
	t, err = time.Parse("1/2/06", input)
	if err == nil {
		year, month, day := t.Date()
		t = time.Date(year, month, day, 23, 59, 59, 0, time.Local)
		return t.Unix()
	}

	t, err = time.Parse("1/2", input)
	if err == nil {
		year, month, day := t.Date()
		currentyear, currentmonth, currentday := time.Now().Date()
		if currentmonth > month || (currentmonth == month && currentday > day) {
			year = currentyear + 1
		} else {
			year = currentyear
		}
		t = time.Date(year, month, day, 23, 59, 59, 0, time.Local)
		return t.Unix()
	}

	return 0
}
