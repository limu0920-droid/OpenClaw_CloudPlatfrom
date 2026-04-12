package httpapi

import "time"

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseRFC3339(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func normalizeRFC3339(value string, fallback string) string {
	if parsed, ok := parseRFC3339(value); ok {
		return parsed.Format(time.RFC3339)
	}
	return fallback
}

func maxRFC3339(values ...string) string {
	var latest time.Time
	for _, value := range values {
		parsed, ok := parseRFC3339(value)
		if !ok {
			continue
		}
		if latest.IsZero() || parsed.After(latest) {
			latest = parsed
		}
	}
	if latest.IsZero() {
		return ""
	}
	return latest.Format(time.RFC3339)
}
