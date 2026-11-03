package circular

// Ground-truth defect: FIX-ARC-02 (Architecture: Circular package dependency / mutual tight coupling)
type ServiceA struct {
	RefB *ServiceB
}

type ServiceB struct {
	RefA *ServiceA
}

func (a *ServiceA) DoA() int {
	if a.RefB != nil {
		return a.RefB.DoB() + 1
	}
	return 1
}

func (b *ServiceB) DoB() int {
	if b.RefA != nil {
		return b.RefA.DoA() + 1
	}
	return 2
}

