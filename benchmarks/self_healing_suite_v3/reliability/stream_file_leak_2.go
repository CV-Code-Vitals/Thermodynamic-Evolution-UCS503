package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessStreamLeak2(paths []string) error {
	for _, p := range paths {
		res, err := os.Open(p)
		if err != nil {
			return err
		}
		_ = res
	}
	return nil
}
