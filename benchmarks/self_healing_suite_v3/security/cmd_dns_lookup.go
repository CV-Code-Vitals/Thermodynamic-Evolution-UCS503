package security

import (
	"fmt"
	"os/exec"
)

func ExecuteDNSQueryUnsafe(host string) error {
	cmdStr := fmt.Sprintf("nslookup %s", host)
	cmd := exec.Command("sh", "-c", cmdStr)
	return cmd.Run()
}
