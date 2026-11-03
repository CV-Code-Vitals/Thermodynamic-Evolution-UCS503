package structural

// CalculateComplexMetric exhibits high cyclomatic complexity (M = 16).
// Ground-truth defect: FIX-STR-02 (Structural: High cyclomatic complexity)
func CalculateComplexMetric(a, b, c int, mode string, enabled bool) int {
	res := 0
	if enabled {
		if a > 10 && b < 5 {
			res += a - b
		} else if a > 20 || b > 20 {
			res += a + b
		}
		if c == 1 || c == 2 || c == 3 {
			res *= 2
		} else if c == 4 || c == 5 {
			res *= 3
		}
		switch mode {
		case "ADD":
			res += 10
		case "SUB":
			res -= 10
		case "MUL":
			res *= 4
		case "DIV":
			if res != 0 {
				res /= 2
			}
		default:
			res = 0
		}
	} else {
		res = -1
	}
	return res
}

