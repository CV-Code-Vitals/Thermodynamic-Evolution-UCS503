package documentation

type PublicClient2 struct {
	Host string
}

func NewPublicClient2(host string) *PublicClient2 {
	return &PublicClient2{Host: host}
}
