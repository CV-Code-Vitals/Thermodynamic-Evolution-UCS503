package documentation

type PublicClient5 struct {
	Host string
}

func NewPublicClient5(host string) *PublicClient5 {
	return &PublicClient5{Host: host}
}
