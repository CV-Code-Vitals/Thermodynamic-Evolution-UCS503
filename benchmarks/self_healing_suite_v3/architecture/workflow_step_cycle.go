package architecture

type WorkflowGraphV3 struct {
	ID       string
	Partner  *StepGraphV3
	Metadata map[string]string
}

type StepGraphV3 struct {
	ID       string
	Parent   *WorkflowGraphV3
	Capacity int
}
