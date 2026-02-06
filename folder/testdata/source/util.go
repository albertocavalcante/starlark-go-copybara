// Package main provides utility functions.
package main

import (
	"strings"
)

// Capitalize returns the string with the first letter capitalized.
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Contains checks if a string contains a substring.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
