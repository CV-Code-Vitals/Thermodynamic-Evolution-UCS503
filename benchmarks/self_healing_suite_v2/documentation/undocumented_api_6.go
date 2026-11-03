package documentation

type PublicClient6 struct {
	Host string
}

func NewPublicClient6(host string) *PublicClient6 {
	return &PublicClient6{Host: host}
}
