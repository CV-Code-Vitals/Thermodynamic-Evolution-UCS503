package structural

// ProcessNestedData exhibits excessive control flow nesting depth (depth = 5).
// Ground-truth defect: FIX-STR-01 (Structural: Excessive nested conditionals)
func ProcessNestedData(data [][]int, threshold int, flag bool) int {
	total := 0
	if flag {
		for _, row := range data {
			if len(row) > 0 {
				for _, val := range row {
					if val > threshold {
						if val%2 == 0 {
							total += val * 2
						} else {
							total += val
						}
					}
				}
			}
		}
	}
	return total
}

