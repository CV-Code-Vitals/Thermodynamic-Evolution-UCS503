package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessDBLeak1(paths []string) error {
	for _, p := range paths {
		res, err := sql.Open("postgres", p)
		if err != nil {
			return err
		}
		_ = res
	}
	return nil
}
