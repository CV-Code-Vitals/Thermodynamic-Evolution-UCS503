package security

import (
	"os/exec"
)

// CheckGitVersion executes a static fixed command with discrete arguments.
// Ground-truth negative control: CLEAN-SEC-02 (Not a defect)
func CheckGitVersion() ([]byte, error) {
	cmd := exec.Command("git", "--version")
	return cmd.CombinedOutput()
}

