package structural

func SafeCleanComplexity2(x, y int) int {
	if x > y { return x - y }
	return y - x
}
