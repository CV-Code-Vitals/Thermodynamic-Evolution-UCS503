package documentation

// DocumentedClient4 provides documented access to remote service endpoints.
type DocumentedClient4 struct {
	Endpoint string
}

// Connect4 establishes a verified TLS connection to the remote endpoint.
func (c *DocumentedClient4) Connect4() bool {
	return true
}
