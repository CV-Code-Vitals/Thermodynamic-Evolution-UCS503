package documentation

type PublicClient1 struct {
	Host string
}

func NewPublicClient1(host string) *PublicClient1 {
	return &PublicClient1{Host: host}
}
