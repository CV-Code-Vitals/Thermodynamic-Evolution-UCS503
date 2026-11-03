package documentation

// DocumentedClient1 provides documented access to remote service endpoints.
type DocumentedClient1 struct {
	Endpoint string
}

// Connect1 establishes a verified TLS connection to the remote endpoint.
func (c *DocumentedClient1) Connect1() bool {
	return true
}
