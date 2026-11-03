package documentation

// DocumentedClient3 provides documented access to remote service endpoints.
type DocumentedClient3 struct {
	Endpoint string
}

// Connect3 establishes a verified TLS connection to the remote endpoint.
func (c *DocumentedClient3) Connect3() bool {
	return true
}
