package api

import (
	"strings"
)

func escapeString(input string) string {
	// must be first
	input = strings.Replace(input, "\\", "\\\\", -1)
	input = strings.Replace(input, ",", "\\,", -1)
	input = strings.Replace(input, "=", "\\=", -1)
	return input
}

func unescapeString(input string) string {
	input = strings.Replace(input, "\\,", ",", -1)
	input = strings.Replace(input, "\\=", "=", -1)
	// must be last.
	input = strings.Replace(input, "\\\\", "\\", -1)
	return input
}
