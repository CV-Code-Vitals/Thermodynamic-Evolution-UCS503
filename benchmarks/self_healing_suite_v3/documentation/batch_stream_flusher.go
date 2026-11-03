package documentation

type BatchStreamFlusherV3 struct {
	Active bool
	Zone   string
}

func (c *BatchStreamFlusherV3) FlushBufferV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
