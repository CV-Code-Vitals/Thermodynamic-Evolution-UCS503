package reliability

import "os"

func DownloadPayloadSafelyClean(p string) ([]byte, error) {
	return os.ReadFile(p)
}
