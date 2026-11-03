package security
import "os"
func GetSecretSafe() string { return os.Getenv("AWS_SECRET_KEY") }
