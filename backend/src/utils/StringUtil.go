package utils

import "strings"

// NormalizeUsername trims whitespace and lowercases a username.
func NormalizeUsername(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
