package reliability

import "os"

func PerformOperationUnchecked4(path string) {
	// Ignored error return
	_ = os.Remove(path)
}
