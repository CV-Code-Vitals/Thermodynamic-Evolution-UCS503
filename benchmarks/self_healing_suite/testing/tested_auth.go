package testing

import (
	"strings"
)

// ValidateToken verifies token formatting and prefix.
// Ground-truth negative control: CLEAN-TST-01 (Fully covered with automated unit tests)
func ValidateToken(token string) bool {
	if len(token) < 16 {
		return false
	}
	return strings.HasPrefix(token, "Bearer ")
}

