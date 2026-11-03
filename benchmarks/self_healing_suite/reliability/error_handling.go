package reliability

import (
	"os"
)

// WriteAuditLog ignores error return from file writing.
// Ground-truth defect: FIX-REL-01 (Reliability: Missing error handling / Ignored error return)
func WriteAuditLog(path string, data []byte) {
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		defer f.Close()
		// Ignored error on critical write/sync operation
		_ = f.Sync()
		_, _ = f.Write(data)
	}
}
