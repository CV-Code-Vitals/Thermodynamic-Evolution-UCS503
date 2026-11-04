package structural

func ComputeMultiDecisionV3_3(a, b, c, d int, mode string) int {
	res := 0
	if a > 0 { res += 1 } else { res -= 1 }
	if b > 10 { res += 2 } else { res -= 2 }
	if c > 20 { res += 3 } else { res -= 3 }
	if d > 30 { res += 4 } else { res -= 4 }
	switch mode {
	case "alpha": res *= 2
	case "beta": res *= 3
	case "gamma": res *= 4
	case "delta": res *= 5
	default: res += 0
	}
	if a == b || c == d { res += 10 }
	if a > c && b > d { res += 20 }
	return res
}
