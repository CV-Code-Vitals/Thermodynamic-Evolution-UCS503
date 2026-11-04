package websocket

import (
	"net"
	"time"
)

// Neg_Control_DialSafe dials socket and defers Close
func DialProbeSafe(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

// INJ-EXT2-RES-03: Resource leak - net.Dial connection opened without Close
func ProbeWebSocketGateway(gatewayAddr string) error {
	conn, err := net.Dial("tcp", gatewayAddr)
	if err != nil {
		return err
	}
	// Missing defer conn.Close() -> RESOURCE_LEAK
	_ = conn
	return nil
}
