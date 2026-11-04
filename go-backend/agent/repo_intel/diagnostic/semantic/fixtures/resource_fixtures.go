package fixtures

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type FileHolder struct {
	F *os.File
}

// 1. Linear os.File leak (Positive)
func Pos_Res_LinearFileLeak(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 2. Deferred os.File.Close (Negative)
func Neg_Res_DeferredFileClose(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// 3. Explicit os.File.Close (Negative)
func Neg_Res_ExplicitFileClose(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	f.Close()
	return data, err
}

// 4. Linear net.Conn leak (Positive)
func Pos_Res_LinearNetConnLeak(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte("ping"))
	return err
}

// 5. Deferred connection close (Negative)
func Neg_Res_DeferredNetConnClose(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte("ping"))
	return err
}

// 6. HTTP response body leak (Positive)
func Pos_Res_HTTPBodyLeak(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

// 7. HTTP response body correctly closed (Negative)
func Neg_Res_HTTPBodyClosed(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

// 8. SQL rows leak (Positive)
func Pos_Res_SQLRowsLeak(db *sql.DB) error {
	rows, err := db.Query("SELECT id FROM users")
	if err != nil {
		return err
	}
	for rows.Next() {
		// do work without rows.Close()
	}
	return rows.Err()
}

// 9. SQL rows deferred close (Negative)
func Neg_Res_SQLRowsDeferredClose(db *sql.DB) error {
	rows, err := db.Query("SELECT id FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		// do work
	}
	return rows.Err()
}

// 10. Transaction with commit (Negative)
func Neg_Res_TxWithCommit(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE accounts SET balance = 100")
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// 11. Transaction with rollback (Negative)
func Neg_Res_TxWithRollback(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	tx.Rollback()
	return nil
}

// 12. Transaction with deferred rollback (Negative)
func Neg_Res_TxWithDeferredRollback(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("INSERT INTO logs VALUES ('done')")
	if err != nil {
		return err
	}
	return tx.Commit()
}

// 13. Conditional resource release (Positive - leaked on else branch)
func Pos_Res_ConditionalRelease(path string, cond bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	if cond {
		f.Close()
		return nil
	}
	return errors.New("aborted") // Leaked along this exit path!
}

// 14. Early-return cleanup (Negative)
func Neg_Res_EarlyReturnCleanup(path string, flag bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if flag {
		return nil
	}
	return errors.New("flag is false")
}

// 15. Error-path cleanup (Negative)
func Neg_Res_ErrorPathCleanup(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Stat(); err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	return nil
}

// 16. Multiple resource aliases (Negative)
func Neg_Res_MultipleAliases(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	aliasF := f
	defer aliasF.Close()
	return nil
}

// 17. Resource returned to caller (Negative)
func Neg_Res_ReturnedToCaller(path string) (*os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// 18. Resource transferred through a supported ownership contract (Negative)
func Neg_Res_TransferredOwnership(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return consumeFile(f)
}

func consumeFile(rc io.ReadCloser) error {
	defer rc.Close()
	return nil
}

// 19. Resource stored in a struct with unknown lifecycle (Positive)
func Pos_Res_StoredInStructUnknownLifecycle(path string) *FileHolder {
	f, _ := os.Open(path)
	return &FileHolder{F: f}
}

// 20. Timer stopped (Negative)
func Neg_Res_TimerStopped(d time.Duration) {
	timer := time.NewTimer(d)
	timer.Stop()
}

// 21. Timer ownership transferred (Negative)
func Neg_Res_TimerTransferred(d time.Duration) *time.Timer {
	timer := time.NewTimer(d)
	return timer
}

// 22. Acquisition failure without resource creation (Negative)
func Neg_Res_AcquisitionFailurePath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		// Acquisition failed, f is nil, no resource obligation
		return err
	}
	defer f.Close()
	return nil
}

// 23. Resource acquired in a loop without closure (Positive)
func Pos_Res_AcquiredInLoop(paths []string) error {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		_ = f
	}
	return nil
}

// 24. Deferred release inside a loop (Positive)
func Pos_Res_DeferredInLoop(paths []string) error {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	return nil
}

// Helper context wrapper
func QueryUser(ctx context.Context, db *sql.DB) (*sql.Rows, error) {
	return db.QueryContext(ctx, "SELECT 1")
}

