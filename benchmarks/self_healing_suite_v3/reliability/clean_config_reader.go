package reliability

import "os"

func ReadConfigFileStrictClean(p string) ([]byte, error) {
	return os.ReadFile(p)
}
