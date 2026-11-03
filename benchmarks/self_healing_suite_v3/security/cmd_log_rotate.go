package security

import (
	"fmt"
	"os/exec"
)

func ExecuteLogRotateUnsafe(host string) error {
	cmdStr := fmt.Sprintf("logrotate %s", host)
	cmd := exec.Command("sh", "-c", cmdStr)
	return cmd.Run()
}
