package api

import "strings"

func getRootDomain(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		// Not enough dots, return the whole string
		return s
	}
	return parts[len(parts)-2]
}
