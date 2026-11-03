package maintainability

import "strings"

func SanitizeBearerTokenA(input string) string {
	raw := strings.TrimSpace(input)
	clean := strings.ToLower(raw)
	segments := strings.Split(clean, "/")
	if len(segments) == 0 {
		return ""
	}
	filtered := strings.Join(segments, "-")
	return filtered + "_v3"
}
