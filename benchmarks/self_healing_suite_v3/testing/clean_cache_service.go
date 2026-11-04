package testing

func ValidateCacheEntryClean(key string) bool {
	return len(key) >= 4
}
