package fixtures

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

// Helper mock types and functions
type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func loadConfig() (*Config, error) {
	return &Config{Host: "localhost", Port: 8080}, nil
}

func fetch() (string, error) {
	return "ok", nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// -------------------------------------------------------------
// POSITIVE FIXTURES (Defects)
// -------------------------------------------------------------

// 1. Unhandled json.Unmarshal error (bare expression)
func Pos_Err_UnhandledUnmarshal(data []byte) map[string]interface{} {
	var res map[string]interface{}
	json.Unmarshal(data, &res)
	return res
}

// 2. Unhandled json.Marshal error where returned error is discarded via blank identifier
func Pos_Err_UnhandledMarshalDiscard(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

// 3. Unhandled file-operation error (os.WriteFile bare expression)
func Pos_Err_UnhandledWriteFile(path string, data []byte) {
	os.WriteFile(path, data, 0644)
}

// 4. Unhandled database-operation error (db.Exec bare expression)
func Pos_Err_UnhandledDBExec(db *sql.DB, query string) {
	db.Exec(query)
}

// 5. Unhandled network-operation error (net.Dial bare expression)
func Pos_Err_UnhandledDial(addr string) {
	net.Dial("tcp", addr)
}

// 6. Explicit _ = operation() for a non-cleanup critical operation
func Pos_Err_ExplicitBlankCriticalOp(data []byte) map[string]interface{} {
	var res map[string]interface{}
	_ = json.Unmarshal(data, &res)
	return res
}

// 7. Multiple-return operation where only the error result is discarded
func Pos_Err_MultipleReturnDiscardError() *Config {
	cfg, _ := loadConfig()
	return cfg
}

// 8. Error overwritten before being checked
func Pos_Err_OverwrittenBeforeCheck(p1, p2 string) error {
	_, err := os.Open(p1)
	_, err = os.Open(p2)
	if err != nil {
		return err
	}
	return nil
}

// 9. Incorrect error-variable check (err2 never checked)
func Pos_Err_MismatchedErrorCheck(v1, v2 interface{}) error {
	_, err1 := json.Marshal(v1)
	_, err2 := json.Marshal(v2)
	if err1 != nil {
		return err1
	}
	_ = err2
	return nil
}

// 10. Error ignored in a branch that continues execution unsafely
func Pos_Err_UnsafeBranchContinue() (string, error) {
	val, err := fetch()
	if err != nil {
		log.Println("fetch failed, ignoring:", err)
		// continues unsafely without returning or handling!
	}
	return val, nil
}

// -------------------------------------------------------------
// NEGATIVE CONTROLS (Safe / Correct Error Handling)
// -------------------------------------------------------------

// 1. Correct if err != nil handling
func Neg_Err_ExplicitIfCheck(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 2. Direct error propagation
func Neg_Err_DirectPropagation(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// 3. Wrapped error propagation
func Neg_Err_WrappedPropagation(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	return data, nil
}

// 4. Correct named return handling
func Neg_Err_NamedReturn(data []byte) (err error) {
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	return
}

// 5. Correct multiple-return handling
func Neg_Err_MultipleReturnHandled() (*Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// 6. Explicitly ignored cleanup error with documented rationale
func Neg_Err_IgnoredCleanup(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_ = f.Close() // Intentional cleanup error suppression
	return nil
}

// 7. Transaction with deferred rollback and commit
func Neg_Err_TxRollbackWithCommit(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return tx.Commit()
}

// 8. Test helper using t.Fatal
func Neg_Err_TestHelperFatal(t *testing.T, data []byte) {
	var res map[string]interface{}
	err := json.Unmarshal(data, &res)
	if err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
}

// 9. Void-returning function without error-producing calls
func Neg_Err_VoidNoErrors() {
	time.Sleep(10 * time.Millisecond)
}

// 10. Error value checked through a recognized helper
func Neg_Err_HelperCheck() (string, error) {
	val, err := fetch()
	check(err)
	return val, nil
}
