package security

import "os/exec"

func ExecuteVersionSafeClean() error {
	cmd := exec.Command("git", "status")
	return cmd.Run()
}
