package security

import "os"

func FetchKeyFromEnvSafeClean() string {
	return os.Getenv("AWS_SECRET_ACCESS_KEY")
}
