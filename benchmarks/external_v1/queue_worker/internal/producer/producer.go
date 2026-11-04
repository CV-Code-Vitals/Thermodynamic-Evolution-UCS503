package producer

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Message represents an event to be queued.
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp int64                  `json:"timestamp"`
}

// Producer sends messages to stream workers.
type Producer struct {
	buffer chan Message
}

// NewProducer creates a buffered message producer.
func NewProducer(bufSize int) *Producer {
	return &Producer{
		buffer: make(chan Message, bufSize),
	}
}

// Publish enqueues a message into the local channel buffer.
func (p *Producer) Publish(msg Message) bool {
	select {
	case p.buffer <- msg:
		return true
	default:
		return false
	}
}

// Out returns the channel of outgoing messages.
func (p *Producer) Out() <-chan Message {
	return p.buffer
}

// SendDirect connects to a remote broker socket and writes raw bytes.
// DEFECT: Network socket resource leak - net.Dial connection is never closed.
func SendDirect(addr string, msg Message) error {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial broker: %w", err)
	}
	// Missing: defer conn.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("socket write failed: %w", err)
	}

	return nil
}

