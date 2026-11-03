package structural

func SafeCleanComplexity1(x, y int) int {
	if x > y { return x - y }
	return y - x
}
