package testing

func ValidatePaymentAuthClean(key string) bool {
	return len(key) >= 4
}
