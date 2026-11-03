package testing

// CalculateInvoiceTax computes complex jurisdictional taxes on high-value billing items
// but has zero unit tests in the package.
// Ground-truth defect: FIX-TST-01 (Testing: Missing tests for critical logic)
func CalculateInvoiceTax(amount float64, state string, exempt bool) float64 {
	if exempt || amount <= 0 {
		return 0.0
	}
	rate := 0.05
	switch state {
	case "NY":
		rate = 0.08875
	case "CA":
		rate = 0.0925
	case "TX":
		rate = 0.0625
	}
	return amount * rate
}

