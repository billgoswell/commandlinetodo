package main

import (
	"fmt"
	"strings"
	"time"
)

func validateListName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("list name cannot be empty")
	}
	const maxListNameLength = 100
	if len(trimmed) > maxListNameLength {
		return "", fmt.Errorf("list name cannot exceed %d characters", maxListNameLength)
	}
	return trimmed, nil
}

func validateTaskText(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", fmt.Errorf("task cannot be empty")
	}
	return trimmed, nil
}

func validatePriority(priority int) int {
	if priority < 1 || priority > 4 {
		return DefaultPriority
	}
	return priority
}

func logErrorMsg(operation string, err error) {
	fmt.Printf("Error: Failed to %s: %v\n", operation, err)
}

func setToEndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, time.Local)
}

func isDateInReasonableRange(t time.Time) bool {
	now := time.Now()
	// Allow dates up to 1 year in the past (for historical tracking)
	if t.Before(now.AddDate(-1, 0, 0)) {
		return false
	}
	// Don't allow dates too far in the future (100 years)
	if t.After(now.AddDate(0, 0, MaxDaysOffset)) {
		return false
	}
	return true
}

func parseDueDate(input string) int64 {
	if input == "" {
		return 0
	}

	var days int
	if len(input) <= 3 {
		_, err := fmt.Sscanf(input, "%d", &days)
		if err == nil && days > 0 && days <= MaxDaysOffset {
			t := time.Now().AddDate(0, 0, days)
			return setToEndOfDay(t).Unix()
		}
	}

	t, err := time.Parse("1/2/2006", input)
	if err == nil && isDateInReasonableRange(t) {
		return setToEndOfDay(t).Unix()
	}
	t, err = time.Parse("1/2/06", input)
	if err == nil && isDateInReasonableRange(t) {
		return setToEndOfDay(t).Unix()
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
		t = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		if isDateInReasonableRange(t) {
			return setToEndOfDay(t).Unix()
		}
	}

	return 0
}
