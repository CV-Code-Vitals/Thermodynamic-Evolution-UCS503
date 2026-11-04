package exec

import (
	"os"
	"os/exec"
)

// Neg_Control_ExecSafe executes command with separated arguments
func RunBackupSafe(archiveName, srcDir string) error {
	cmd := exec.Command("tar", "-czf", archiveName, srcDir)
	return cmd.Run()
}

// DEF-HOLDOUT-04: Shell command injection via sh -c with string concatenation
func DispatchShellTask(topic string) error {
	cmd := exec.Command("sh", "-c", "echo task for: "+topic)
	return cmd.Run()
}

// DEF-HOLDOUT-06: Unhandled error on os.WriteFile
func DumpTelemetryLog(path string, data []byte) {
	os.WriteFile(path, data, 0644) // UNHANDLED_ERROR: return value ignored
}

