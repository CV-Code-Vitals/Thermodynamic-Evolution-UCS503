package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessDeferLoop2(items []string) error {
	for _, p := range items {
		res, err := os.Open(p)
		if err != nil {
			return err
		}
		defer res.Close()
	}
	return nil
}
