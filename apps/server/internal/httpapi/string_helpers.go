package httpapi

import "strings"

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
