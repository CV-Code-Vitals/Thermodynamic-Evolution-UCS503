package documentation

type RouterDispatcherV3 struct {
	Active bool
	Zone   string
}

func (c *RouterDispatcherV3) DispatchEventV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
