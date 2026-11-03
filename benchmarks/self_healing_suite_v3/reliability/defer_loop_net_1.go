package reliability

import (
	"database/sql"
	"net"
	"os"
)

func ProcessDeferLoopNet1(items []string) error {
	for _, p := range items {
		res, err := net.Dial("udp", p)
		if err != nil {
			return err
		}
		defer res.Close()
	}
	return nil
}
