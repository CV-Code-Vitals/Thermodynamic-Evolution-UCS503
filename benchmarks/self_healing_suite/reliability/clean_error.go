package reliability

import (
	"errors"
	"os"
)

// ReadConfigWithFallback reads a config file with robust error handling and fallback default.
// Ground-truth negative control: CLEAN-REL-01 (Not a defect)
func ReadConfigWithFallback(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.New("empty path provided")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte("default: true\n"), nil
		}
		return nil, err
	}
	return data, nil
}

