package pkg_c

// DataProcessor defines single-target interface (CP-IFACE-SINGLE-06).
type DataProcessor interface {
	ProcessData(data []byte) int
}

// FastProcessor is the only implementation of DataProcessor.
type FastProcessor struct{}

func (f *FastProcessor) ProcessData(data []byte) int {
	return len(data)
}

// MultiHandler defines multiple-target interface (CP-IFACE-MULTI-07).
type MultiHandler interface {
	Handle(val string) string
}

type HandlerOne struct{}

func (h *HandlerOne) Handle(val string) string {
	return "one:" + val
}

type HandlerTwo struct{}

func (h *HandlerTwo) Handle(val string) string {
	return "two:" + val
}

// Factorial tests direct recursion handling and termination (CP-REC-DIRECT-08).
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// Ping tests mutual recursion (CP-REC-MUTUAL-09).
func Ping(n int) int {
	if n <= 0 {
		return 0
	}
	return Pong(n - 1)
}

// Pong completes mutual recursion (CP-REC-MUTUAL-09).
func Pong(n int) int {
	if n <= 0 {
		return 0
	}
	return Ping(n - 1)
}

// DepthStep4 is at depth K=4 (CP-DEPTH-K-10).
func DepthStep4(val int) int {
	return val * 4
}

// DepthStep3 is at depth K=3.
func DepthStep3(val int) int {
	return DepthStep4(val + 3)
}

// DepthStep2 is at depth K=2.
func DepthStep2(val int) int {
	return DepthStep3(val + 2)
}

// DepthStep1 is caller initiating depth chain (depth K=1).
func DepthStep1(val int) int {
	return DepthStep2(val + 1)
}

// RunDynamicDispatch tests dynamic dispatch on interface parameter (CP-DYN-DISP-17).
func RunDynamicDispatch(p DataProcessor, input []byte) int {
	return p.ProcessData(input)
}

// CleanInterfaceControl provides negative control for interface invocation.
func CleanInterfaceControl() int {
	proc := &FastProcessor{}
	return proc.ProcessData([]byte("hello"))
}

