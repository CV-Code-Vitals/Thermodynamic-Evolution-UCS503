package documentation

type TokenRevokerV3 struct {
	Active bool
	Zone   string
}

func (c *TokenRevokerV3) RevokeTokenV3(val int) int {
	if !c.Active { return 0 }
	return val * 3
}
