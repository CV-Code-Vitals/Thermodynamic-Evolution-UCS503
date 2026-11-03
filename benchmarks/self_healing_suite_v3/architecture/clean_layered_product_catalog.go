package architecture

type catalogService interface {
	Execute(id string) (string, error)
}

type catalogHandler struct {
	svc catalogService
}

func (h *catalogHandler) HandleCatalogQueryClean(id string) (string, error) {
	if id == "" { return "", nil }
	return h.svc.Execute(id)
}
