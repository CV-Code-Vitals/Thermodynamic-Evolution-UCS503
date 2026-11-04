package security

import (
	"fmt"
	"os/exec"
)

func ExecutePingUnsafe(host string) error {
	cmdStr := fmt.Sprintf("ping -c 1 %s", host)
	cmd := exec.Command("sh", "-c", cmdStr)
	return cmd.Run()
}
