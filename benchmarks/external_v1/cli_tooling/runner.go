package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// ExecuteDirect executes a binary directly without shell interpretation (clean implementation).
func ExecuteDirect(ctx context.Context, binary string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("command %s failed: %w", binary, err)
	}
	return string(out), nil
}

// ExecuteShellScript runs a command through the shell interpreter.
// DEFECT: Command injection vulnerability - string concatenated shell command executed via shell interpreter.
func ExecuteShellScript(userInput string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd.exe", "/C", userInput)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", userInput)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("shell execution failed: %w", err)
	}
	return string(out), nil
}

