package architecture

type userService interface {
	Execute(id string) (string, error)
}

type userHandler struct {
	svc userService
}

func (h *userHandler) HandleUserQueryClean(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
