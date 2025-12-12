package main

import (
	"testing"
	"time"
)

func TestParseDueDate_EmptyString(t *testing.T) {
	result := parseDueDate("")
	if result != 0 {
		t.Errorf("expected 0 for empty string, got %d", result)
	}
}

func TestParseDueDate_InvalidInput(t *testing.T) {
	tests := []string{
		"invalid",
		"abc",
		"13/32/2025",
		"-1",
		"0",
	}
	for _, input := range tests {
		result := parseDueDate(input)
		if result != 0 {
			t.Errorf("expected 0 for invalid input %q, got %d", input, result)
		}
	}
}

func TestParseDueDate_DaysFromNow(t *testing.T) {
	tests := []struct {
		input string
		days  int
	}{
		{"1", 1},
		{"3", 3},
		{"7", 7},
		{"30", 30},
		{"999", 999},
	}

	for _, tt := range tests {
		result := parseDueDate(tt.input)
		if result == 0 {
			t.Errorf("expected non-zero result for %q days", tt.input)
			continue
		}

		resultTime := time.Unix(result, 0)
		expected := time.Now().AddDate(0, 0, tt.days)

		if resultTime.Year() != expected.Year() ||
			resultTime.Month() != expected.Month() ||
			resultTime.Day() != expected.Day() {
			t.Errorf("for input %q: expected date %v, got %v",
				tt.input, expected.Format("2006-01-02"), resultTime.Format("2006-01-02"))
		}

		if resultTime.Hour() != 23 || resultTime.Minute() != 59 || resultTime.Second() != 59 {
			t.Errorf("for input %q: expected time 23:59:59, got %v",
				tt.input, resultTime.Format("15:04:05"))
		}
	}
}

func TestParseDueDate_FullDateFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"12/25/2025", "2025-12-25"},
		{"1/1/2026", "2026-01-01"},
		{"6/15/2025", "2025-06-15"},
	}

	for _, tt := range tests {
		result := parseDueDate(tt.input)
		if result == 0 {
			t.Errorf("expected non-zero result for %q", tt.input)
			continue
		}

		resultTime := time.Unix(result, 0)
		resultDate := resultTime.Format("2006-01-02")
		if resultDate != tt.expected {
			t.Errorf("for input %q: expected %s, got %s", tt.input, tt.expected, resultDate)
		}
	}
}

func TestParseDueDate_ShortYearFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"12/25/25", "2025-12-25"},
		{"1/1/26", "2026-01-01"},
		{"6/15/30", "2030-06-15"},
	}

	for _, tt := range tests {
		result := parseDueDate(tt.input)
		if result == 0 {
			t.Errorf("expected non-zero result for %q", tt.input)
			continue
		}

		resultTime := time.Unix(result, 0)
		resultDate := resultTime.Format("2006-01-02")
		if resultDate != tt.expected {
			t.Errorf("for input %q: expected %s, got %s", tt.input, tt.expected, resultDate)
		}
	}
}

func TestParseDueDate_MonthDayOnly(t *testing.T) {
	currentYear := time.Now().Year()

	result := parseDueDate("6/15")
	if result == 0 {
		t.Error("expected non-zero result for month/day format")
		return
	}

	resultTime := time.Unix(result, 0)
	resultYear := resultTime.Year()

	if resultYear != currentYear && resultYear != currentYear+1 {
		t.Errorf("expected year to be %d or %d, got %d", currentYear, currentYear+1, resultYear)
	}
}

func TestParseDueDate_TimeSetToEndOfDay(t *testing.T) {
	inputs := []string{"1", "12/25/2025", "12/25/25", "6/15"}

	for _, input := range inputs {
		result := parseDueDate(input)
		if result == 0 {
			continue
		}

		resultTime := time.Unix(result, 0)
		if resultTime.Hour() != 23 || resultTime.Minute() != 59 || resultTime.Second() != 59 {
			t.Errorf("for input %q: expected time 23:59:59, got %s",
				input, resultTime.Format("15:04:05"))
		}
	}
}
