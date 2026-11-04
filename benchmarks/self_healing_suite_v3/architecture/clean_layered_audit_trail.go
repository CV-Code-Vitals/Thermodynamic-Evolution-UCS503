package architecture

type auditService interface {
	Execute(id string) (string, error)
}

type auditHandler struct {
	svc auditService
}

func (h *auditHandler) HandleAuditQueryClean(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
