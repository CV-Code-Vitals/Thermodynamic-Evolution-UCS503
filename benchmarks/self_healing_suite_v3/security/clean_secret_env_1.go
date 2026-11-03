package security

import "os"

func FetchTokenFromEnvSafeClean() string {
	return os.Getenv("AWS_SECRET_ACCESS_KEY")
}
