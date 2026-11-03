package documentation

type CircuitBreakerPanelV3 struct {
	Active bool
	Zone   string
}

func (c *CircuitBreakerPanelV3) RecordTripV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
