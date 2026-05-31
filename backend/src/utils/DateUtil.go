package utils

import "time"

// NowISO returns the current UTC time formatted as RFC3339.
func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// FormatISO formats a time.Time value as RFC3339 UTC.
func FormatISO(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
