package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessSocketLeak1(paths []string) error {
	for _, p := range paths {
		res, err := net.Dial("tcp", p)
		if err != nil {
			return err
		}
		_ = res
	}
	return nil
}
