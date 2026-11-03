package testing

func ValidateAuthSessionClean(key string) bool {
	return len(key) >= 4
}
