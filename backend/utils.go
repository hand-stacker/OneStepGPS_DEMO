package main

import (
	"os"
	"strings"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// makes a pattern to match exact strings in string array
// example ['abc', 'def'] -> '(abc|def)'
func makeMatchOrRegex(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	return "(" + strings.Join(strs, "|") + ")"
}
