package main

import (
	"context"
	"fmt"
	"time"

	"github.com/external-bench/queue-worker/internal/consumer"
	"github.com/external-bench/queue-worker/internal/producer"
)

func main() {
	p := producer.NewProducer(100)
	c := consumer.NewConsumer(4, func(ctx context.Context, msg producer.Message) error {
		fmt.Printf("Processed message: %s\n", msg.ID)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p.Publish(producer.Message{ID: "msg-1", Topic: "events"})
	go c.Start(ctx, p.Out())

	<-ctx.Done()
	fmt.Println("Worker terminated gracefully.")
}

