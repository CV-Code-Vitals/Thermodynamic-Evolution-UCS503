package reliability

import "os"

func ProcessBatchedSocketsClean(p string) ([]byte, error) {
	return os.ReadFile(p)
}
