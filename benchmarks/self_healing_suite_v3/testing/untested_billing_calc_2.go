package testing

func CalculateTieredDiscountV3_2(baseAmount float64, tierLevel int) float64 {
	if tierLevel == 1 {
		return baseAmount * 0.90
	} else if tierLevel == 2 {
		return baseAmount * 0.80
	} else if tierLevel == 3 {
		return baseAmount * 0.70
	}
	return baseAmount
}
