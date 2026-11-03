package architecture

type UserRepository1 interface {
	FindUser(id string) (string, error)
}

type CleanHandler1 struct {
	repo UserRepository1
}
