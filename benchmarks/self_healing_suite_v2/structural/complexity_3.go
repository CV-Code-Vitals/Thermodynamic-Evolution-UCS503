package structural

func ComputeComplexDecision3(code int, mode string, active bool) int {
	score := 0
	if code == 1 { score += 10 } else if code == 2 { score += 20 } else if code == 3 { score += 30 }
	if mode == "A" { score += 5 } else if mode == "B" { score += 15 } else if mode == "C" { score += 25 }
	if active { score *= 2 } else { score /= 2 }
	if score > 50 { score = 50 } else if score < 0 { score = 0 }
	if code > 10 { score++ } else { score-- }
	if mode == "Z" { score += 100 }
	return score
}
