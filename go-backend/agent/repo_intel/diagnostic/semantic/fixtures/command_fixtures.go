package fixtures

import (
	"context"
	"fmt"
	"log"
	"os/exec"
)

// POSITIVE FIXTURES: Vulnerable command execution patterns

// Pos_Cmd_ShellSprintf demonstrates untrusted input interpolated into shell command via fmt.Sprintf.
func Pos_Cmd_ShellSprintf(userInput string) error {
	cmdStr := fmt.Sprintf("grep %s /var/log/audit.log", userInput)
	return exec.Command("sh", "-c", cmdStr).Run()
}

// Pos_Cmd_ShellConcat demonstrates untrusted input concatenated into bash -c.
func Pos_Cmd_ShellConcat(userInput string) ([]byte, error) {
	return exec.Command("bash", "-c", "echo "+userInput).CombinedOutput()
}

// Pos_Cmd_ShellPropagated demonstrates untrusted input propagated through local variables into cmd.exe /c.
func Pos_Cmd_ShellPropagated(script string) error {
	raw := fmt.Sprintf("dir %s", script)
	target := raw
	return exec.Command("cmd.exe", "/c", target).Run()
}

// Pos_Cmd_ShellParamDirect demonstrates passing a caller-controlled command string directly to sh -c.
func Pos_Cmd_ShellParamDirect(commandString string) error {
	return exec.Command("sh", "-c", commandString).Run()
}

// Pos_Cmd_ContextShell demonstrates dynamic command execution under a context-aware sink.
func Pos_Cmd_ContextShell(ctx context.Context, userInput string) error {
	cmd := fmt.Sprintf("cat %s", userInput)
	return exec.CommandContext(ctx, "sh", "-c", cmd).Run()
}

// NEGATIVE FIXTURES: Safe controls and non-vulnerable command patterns

// Neg_Cmd_StaticExec demonstrates constant executable and arguments.
func Neg_Cmd_StaticExec() error {
	return exec.Command("git", "status", "--short").Run()
}

// Neg_Cmd_SafeArgSeparation demonstrates safe argument separation with dynamic user input.
func Neg_Cmd_SafeArgSeparation(userInput string) error {
	return exec.Command("git", "log", "--grep", userInput).Run()
}

// Neg_Cmd_DirectArgBranch demonstrates passing dynamic arguments without shell interpretation.
func Neg_Cmd_DirectArgBranch(branch string) error {
	return exec.Command("git", "checkout", branch).Run()
}

// Neg_Cmd_StaticShell demonstrates static shell execution with constant string.
func Neg_Cmd_StaticShell() error {
	return exec.Command("sh", "-c", "echo static safe text").Run()
}

// Neg_Cmd_SprintfNonExec demonstrates fmt.Sprintf used outside of command sinks.
func Neg_Cmd_SprintfNonExec(userInput string) {
	msg := fmt.Sprintf("Processed command request for user %s", userInput)
	log.Println(msg)
}

// Neg_Cmd_UnusedCommand demonstrates dynamic command string construction that never reaches an execution sink.
func Neg_Cmd_UnusedCommand(userInput string) string {
	unusedCmd := fmt.Sprintf("tar -czf backup.tar.gz %s", userInput)
	return unusedCmd
}
