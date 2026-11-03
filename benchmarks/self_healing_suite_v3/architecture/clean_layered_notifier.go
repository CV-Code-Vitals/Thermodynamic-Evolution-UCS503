package architecture

type notifierService interface {
	Execute(id string) (string, error)
}

type notifierHandler struct {
	svc notifierService
}

func (h *notifierHandler) HandleNotificationQueryClean(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
