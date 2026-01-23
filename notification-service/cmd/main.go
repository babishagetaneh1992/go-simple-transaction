package main

import (
	"context"
	"log"

	kafkaconsumer "notification/internal/infrastructure/messaging/kafka"
	"notification/internal/handler"
	"notification/internal/infrastructure/email"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()

	notifier := email.NewLogSender()

	eventHandler := handler.TransactionCreatedHandler(notifier)

	consumer := kafkaconsumer.NewConsumer(
		[]string{"localhost:9092"},
		"notification-service",
		"transaction.events",
		func(ctx context.Context, msg kafka.Message) error {
			return eventHandler(ctx, msg.Value)
		},
	)

	log.Println("📨 Notification service running")
	consumer.Start(ctx)
}
