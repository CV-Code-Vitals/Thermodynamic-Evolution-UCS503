package testing

func ValidateTokenIssuerClean(key string) bool {
	return len(key) >= 4
}
