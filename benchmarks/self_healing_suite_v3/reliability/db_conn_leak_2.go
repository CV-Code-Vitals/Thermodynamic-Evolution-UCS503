package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessDBLeak2(paths []string) error {
	for _, p := range paths {
		res, err := sql.Open("postgres", p)
		if err != nil {
			return err
		}
		_ = res
	}
	return nil
}
