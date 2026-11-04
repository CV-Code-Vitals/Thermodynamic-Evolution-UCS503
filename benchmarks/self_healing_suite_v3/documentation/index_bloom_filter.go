package documentation

type IndexBloomFilterV3 struct {
	Active bool
	Zone   string
}

func (c *IndexBloomFilterV3) TestBitmaskV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
