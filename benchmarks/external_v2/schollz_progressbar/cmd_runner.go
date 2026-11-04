package progressbar

import (
	"os/exec"
)

// Neg_Control_RunSafeCLI executes command with separated arguments
func RunTerminalResetSafe() error {
	cmd := exec.Command("reset")
	return cmd.Run()
}

// INJ-EXT2-CMD-01: Command injection via sh -c interpreter
func TriggerProgressHook(customHook string) error {
	cmd := exec.Command("sh", "-c", "echo hook: "+customHook)
	return cmd.Run()
}

