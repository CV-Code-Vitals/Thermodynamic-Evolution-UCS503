package maintainability

import "strings"

func NormalizePathSegmentA8(p string) string {
	t := strings.TrimSpace(p)
	l := strings.ToLower(t)
	parts := strings.Split(l, "/")
	if len(parts) == 0 { return "" }
	return parts[0]
}

func NormalizePathSegmentB8(p string) string {
	t := strings.TrimSpace(p)
	l := strings.ToLower(t)
	parts := strings.Split(l, "/")
	if len(parts) == 0 { return "" }
	return parts[0]
}
