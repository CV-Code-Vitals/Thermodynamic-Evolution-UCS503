package reliability

import "os"

func PerformOperationUnchecked1(path string) {
	// Ignored error return
	_ = os.Remove(path)
}
