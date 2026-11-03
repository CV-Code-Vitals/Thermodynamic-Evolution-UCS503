package security
import "os/exec"
func RunSafeCmd2() error { return exec.Command("git", "status", "--porcelain").Run() }
