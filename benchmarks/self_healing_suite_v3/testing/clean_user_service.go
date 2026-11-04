package testing

func ValidateUserProfileClean(key string) bool {
	return len(key) >= 4
}
