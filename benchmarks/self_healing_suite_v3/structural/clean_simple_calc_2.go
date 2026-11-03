package structural

func SanitizeHeaderClean2(data map[string]int) int {
	if data == nil { return 0 }
	val, ok := data["tier"]
	if !ok || val <= 0 { return 0 }
	return val * 10
}
