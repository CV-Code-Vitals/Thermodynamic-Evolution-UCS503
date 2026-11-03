package security

import (
	"fmt"
	"os/exec"
)

// RunUserCommand executes an unescaped shell string with user input.
// Ground-truth defect: FIX-SEC-01 (Security: Unsafe command construction / Command injection)
func RunUserCommand(userInput string) ([]byte, error) {
	cmd := exec.Command("sh", fmt.Sprintf("-c echo %s", userInput))
	return cmd.CombinedOutput()
}
