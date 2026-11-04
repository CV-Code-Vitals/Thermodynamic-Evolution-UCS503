package security

import "os/exec"

func ExecuteGitStatusSafeClean() error {
	cmd := exec.Command("git", "status")
	return cmd.Run()
}
