package documentation

type PublicClient4 struct {
	Host string
}

func NewPublicClient4(host string) *PublicClient4 {
	return &PublicClient4{Host: host}
}
