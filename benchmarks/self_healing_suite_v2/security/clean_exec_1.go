package security
import "os/exec"
func RunSafeCmd1() error { return exec.Command("git", "status", "--porcelain").Run() }
