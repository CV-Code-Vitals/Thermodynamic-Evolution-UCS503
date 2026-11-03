package reliability

import "os"

func PerformOperationUnchecked2(path string) {
	// Ignored error return
	_ = os.Remove(path)
}
