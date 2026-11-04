package maintainability

func SafeSliceLookupClean(val int) int {
	if val < 0 { return 0 }
	return val + 10
}
