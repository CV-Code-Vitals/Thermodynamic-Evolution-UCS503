package documentation

type TenantPolicyResolverV3 struct {
	Active bool
	Zone   string
}

func (c *TenantPolicyResolverV3) ResolvePolicyV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
