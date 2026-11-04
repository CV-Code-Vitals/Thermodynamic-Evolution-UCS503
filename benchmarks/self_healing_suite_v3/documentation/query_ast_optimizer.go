package documentation

type QueryAstOptimizerV3 struct {
	Active bool
	Zone   string
}

func (c *QueryAstOptimizerV3) RewritePlanV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
