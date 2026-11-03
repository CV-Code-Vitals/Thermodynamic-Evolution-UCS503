package reliability

import "os"

func SafeResourceCleanup2(p string) ([]byte, error) {
	data, err := os.ReadFile(p)
	if err != nil { return nil, err }
	return data, nil
}
