package architecture

type UserRepository2 interface {
	FindUser(id string) (string, error)
}

type CleanHandler2 struct {
	repo UserRepository2
}
