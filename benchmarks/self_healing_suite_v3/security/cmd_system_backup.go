package security

import (
	"fmt"
	"os/exec"
)

func ExecuteBackupUnsafe(host string) error {
	cmdStr := fmt.Sprintf("tar -czf %s", host)
	cmd := exec.Command("sh", "-c", cmdStr)
	return cmd.Run()
}
