package pkg_b

import (
	"errors"
	"fmt"
	"os"
)

// Service encapsulates state for method receiver testing.
type Service struct {
	Name string
}

// Process tests value receiver resolution.
func (s Service) Process(input string) string {
	return s.Name + ":" + input
}

// Mutate tests pointer receiver resolution.
func (s *Service) Mutate(newName string) {
	s.Name = newName
}

// FormatQuery constructs an unsanitized SQL query across package boundary.
func FormatQuery(table, col, val string) string {
	return fmt.Sprintf("SELECT * FROM %s WHERE %s = '%s'", table, col, val)
}

// BuildShellScript constructs an unsanitized shell command string across package boundary.
func BuildShellScript(param string) string {
	return "echo 'Running: ' && " + param
}

// AcquireFile acquires an OS file and transfers ownership to the caller.
func AcquireFile(path string) (*os.File, error) {
	return os.Open(path)
}

// CloserHelper takes ownership and closes a resource in a callee function.
func CloserHelper(f *os.File) error {
	if f != nil {
		return f.Close()
	}
	return nil
}

// WrapError wraps an incoming error.
func WrapError(err error) error {
	if err != nil {
		return fmt.Errorf("wrapped: %w", err)
	}
	return nil
}

// NamedReturnCalc demonstrates named error return across package boundary.
func NamedReturnCalc(x int) (res int, err error) {
	if x < 0 {
		err = errors.New("negative value not allowed")
		return
	}
	res = x * 2
	return
}

