//go:build linux
// +build linux

package pkg_d

// PlatformSpecificAction is only valid under linux build tag (CP-BUILD-TAG-19).
func PlatformSpecificAction() string {
	return "linux_specific_execution"
}

