package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
)

var validSortTypes = []string{"asc", "desc"}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// makes a pattern to match exact strings in string array
// example ['abc', 'def'] -> '(abc|def)'
func makeMatchOrRegex(strs []string) string {
	// empty strings will match all ids.. if they are or'ed : "(|apple|banana)"
	var nonEmpty []string
	for _, s := range strs {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	return "(" + strings.Join(nonEmpty, "|") + ")"
}

// parses /user_id/ (from a url) into int 64
func getInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func validSortType(s string) bool {
	for _, v := range validSortTypes {
		if v == s {
			return true
		}
	}
	return false
}

// /// HTTP USEFUL ERROR RETURNS
func throwBadRequest(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

func throwStatusNotFound(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusNotFound)
}

func throwInternalServerError(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusInternalServerError)
}
