package documentation

type PublicClient3 struct {
	Host string
}

func NewPublicClient3(host string) *PublicClient3 {
	return &PublicClient3{Host: host}
}
