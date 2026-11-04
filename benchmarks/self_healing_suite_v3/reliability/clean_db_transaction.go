package reliability

import "os"

func TransactDatabaseSafeClean(p string) ([]byte, error) {
	return os.ReadFile(p)
}
