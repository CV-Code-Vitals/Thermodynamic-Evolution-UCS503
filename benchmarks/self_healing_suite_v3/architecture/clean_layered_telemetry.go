package architecture

type telemetryService interface {
	Execute(id string) (string, error)
}

type telemetryHandler struct {
	svc telemetryService
}

func (h *telemetryHandler) HandleTelemetryQueryClean(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
