package execution

import (
	"fmt"
	"os"
	"path/filepath"
)

// RollbackEngine manages restoring isolated workspaces to pristine baseline state.
type RollbackEngine struct {
	OriginalRepoRoot string
}

// NewRollbackEngine creates a new RollbackEngine.
func NewRollbackEngine(originalRepoRoot string) *RollbackEngine {
	return &RollbackEngine{
		OriginalRepoRoot: originalRepoRoot,
	}
}

// RollbackWorkspace restores modified files in the workspace by copying original versions back.
func (r *RollbackEngine) RollbackWorkspace(wsDir string, modifiedFiles []string, reason string) error {
	for _, relPath := range modifiedFiles {
		origPath := filepath.Join(r.OriginalRepoRoot, filepath.FromSlash(relPath))
		wsPath := filepath.Join(wsDir, filepath.FromSlash(relPath))

		// If the file existed originally, copy original content back
		if origInfo, err := os.Stat(origPath); err == nil && !origInfo.IsDir() {
			if err := copyFile(origPath, wsPath); err != nil {
				return fmt.Errorf("rollback failed restoring %s: %w", wsPath, err)
			}
		} else if os.IsNotExist(err) {
			// If file did not exist originally (was newly created), remove it
			if err := os.Remove(wsPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("rollback failed removing newly created %s: %w", wsPath, err)
			}
		}
	}

	return nil
}

