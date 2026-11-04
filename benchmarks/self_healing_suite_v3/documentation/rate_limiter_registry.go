package documentation

type RateLimiterRegistryV3 struct {
	Active bool
	Zone   string
}

func (c *RateLimiterRegistryV3) CheckQuotaV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
