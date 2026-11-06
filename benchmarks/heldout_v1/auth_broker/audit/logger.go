package audit

import (
	"fmt"
	"os"
)

// Neg_Control_LogEventSafe opens file and defers Close
func LogEventSafe(logPath, event string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, event)
	return err
}

// DEF-HOLDOUT-09: Resource leak - file handle opened without Close
func RotateAuditLog(logPath, archivePath string) error {
	f, err := os.OpenFile(logPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	// Missing defer f.Close() -> RESOURCE_LEAK
	_ = archivePath
	return nil
}

