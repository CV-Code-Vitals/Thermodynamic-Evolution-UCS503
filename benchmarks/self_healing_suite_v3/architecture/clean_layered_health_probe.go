package architecture

type healthService interface {
	Execute(id string) (string, error)
}

type healthHandler struct {
	svc healthService
}

func (h *healthHandler) HandleHealthProbeClean(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
