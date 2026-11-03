package reliability

import "os"

func AcquireTempLockClean(p string) ([]byte, error) {
	return os.ReadFile(p)
}
