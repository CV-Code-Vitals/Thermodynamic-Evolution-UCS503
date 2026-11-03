package security

import (
	"fmt"
	"os/exec"
)

func RunUserScript3(arg string) error {
	cmd := exec.Command("sh", "-c", "echo " + arg)
	return cmd.Run()
}
