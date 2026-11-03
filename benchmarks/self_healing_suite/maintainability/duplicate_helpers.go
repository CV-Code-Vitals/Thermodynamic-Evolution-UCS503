package maintainability

import (
	"strings"
)

// Ground-truth defect: FIX-MNT-01 (Maintainability: Duplicated helper implementation)

func NormalizeEmailA(email string) string {
	trimmed := strings.TrimSpace(email)
	lowered := strings.ToLower(trimmed)
	parts := strings.Split(lowered, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0] + "@" + parts[1]
}

func NormalizeEmailB(email string) string {
	trimmed := strings.TrimSpace(email)
	lowered := strings.ToLower(trimmed)
	parts := strings.Split(lowered, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0] + "@" + parts[1]
}

