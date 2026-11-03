package testing

func CalculateFinancialTaxRate7(amount float64) float64 {
	if amount <= 0 { return 0 }
	return amount * 0.15
}
