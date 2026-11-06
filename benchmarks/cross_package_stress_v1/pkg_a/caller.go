package pkg_a

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/exec"

	"cross_pkg_stress/pkg_b"
	b_alias "cross_pkg_stress/pkg_b"
)

// CallerRunCrossFile tests same-package cross-file invocation (CP-SAME-PKG-01).
func CallerRunCrossFile(input string) string {
	return LocalFormat(input)
}

// CallerRunImported tests imported package invocation (CP-IMP-FUNC-02).
func CallerRunImported(table, col, val string) string {
	return pkg_b.FormatQuery(table, col, val)
}

// CallerRunAlias tests package alias invocation (CP-PKG-ALIAS-03).
func CallerRunAlias(param string) string {
	return b_alias.BuildShellScript(param)
}

// CallerRunReceiver tests value and pointer receiver dispatch (CP-RECV-METHOD-04, CP-PTR-RECV-05).
func CallerRunReceiver() (string, string) {
	svc := pkg_b.Service{Name: "Alpha"}
	valRes := svc.Process("test")

	svcPtr := &pkg_b.Service{Name: "Beta"}
	svcPtr.Mutate("Gamma")
	ptrRes := svcPtr.Process("test")

	return valRes, ptrRes
}

// CallerSQLCrossPackage tests whether SQL taint constructed in callee propagates into caller sink (CP-SQL-HELPER-15).
func CallerSQLCrossPackage(ctx context.Context, db *sql.DB, user string) error {
	query := pkg_b.FormatQuery("users", "name", user)
	_, err := db.QueryContext(ctx, query)
	return err
}

// CallerCommandCrossPackage tests whether shell script constructed in aliased callee propagates into caller sink (CP-CMD-HELPER-16).
func CallerCommandCrossPackage(ctx context.Context, script string) error {
	shellCmd := b_alias.BuildShellScript(script)
	cmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	return cmd.Run()
}

// CallerResourceTransfer tests cross-package resource acquisition and ownership transfer (CP-RES-TRANS-14).
func CallerResourceTransfer(path string) error {
	f, err := pkg_b.AcquireFile(path)
	if err != nil {
		return err
	}
	// Transferred resource is never closed in caller!
	_ = f.Name()
	return nil
}

// CallerResourceSplit tests resource acquisition locally with cleanup in callee (CP-RES-SPLIT-13).
func CallerResourceSplit(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	// Closed in helper function across package boundary
	return pkg_b.CloserHelper(f)
}

// CallerErrorPropagation tests error wrapper handling across packages (CP-ERR-WRAP-11).
func CallerErrorPropagation() {
	origErr := errors.New("upstream failure")
	wrapped := pkg_b.WrapError(origErr)
	_ = wrapped // Silently ignored wrapped error
}

// CallerNamedReturn tests named return values across packages (CP-NAMED-RET-12).
func CallerNamedReturn() (int, error) {
	res, err := pkg_b.NamedReturnCalc(10)
	if err != nil {
		return 0, err
	}
	return res, nil
}

