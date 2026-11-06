package pkg_a

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
)

// LocalFormat tests same-package cross-file resolution.
func LocalFormat(param string) string {
	return "prefix_" + param
}

// SafeQuery provides a negative clean control for SQL execution.
func SafeQuery(ctx context.Context, db *sql.DB, id string) error {
	_, err := db.QueryContext(ctx, "SELECT id, name FROM users WHERE id = $1", id)
	return err
}

// SafeCommand provides a negative clean control for command execution.
func SafeCommand(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "ls", "-la")
	return cmd.Run()
}

// SafeResource provides a negative clean control for resource lifecycle.
func SafeResource(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

