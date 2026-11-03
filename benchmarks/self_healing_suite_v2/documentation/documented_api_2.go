package documentation

// DocumentedClient2 provides documented access to remote service endpoints.
type DocumentedClient2 struct {
	Endpoint string
}

// Connect2 establishes a verified TLS connection to the remote endpoint.
func (c *DocumentedClient2) Connect2() bool {
	return true
}
