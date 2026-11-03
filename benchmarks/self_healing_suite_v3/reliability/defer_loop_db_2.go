package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessDeferLoopDB2(items []string) error {
	for _, p := range items {
		res, err := sql.Open("sqlite3", p)
		if err != nil {
			return err
		}
		defer res.Close()
	}
	return nil
}
