package reliability

import "os"

func PerformOperationUnchecked3(path string) {
	// Ignored error return
	_ = os.Remove(path)
}
