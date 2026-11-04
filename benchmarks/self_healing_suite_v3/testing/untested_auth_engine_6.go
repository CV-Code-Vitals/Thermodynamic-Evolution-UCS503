package testing

func EvaluateTokenScopesV3_6(scopes []string, targetScope string) bool {
	for _, s := range scopes {
		if s == targetScope || s == "admin" {
			return true
		}
	}
	return false
}
