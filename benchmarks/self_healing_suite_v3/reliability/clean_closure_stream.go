package reliability

import "os"

func StreamFileWithClosureClean(p string) ([]byte, error) {
	return os.ReadFile(p)
}
