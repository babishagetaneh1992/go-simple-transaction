package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	handler func(context.Context, kafka.Message) error
}


func NewConsumer(
	brokers []string,
	groupID  string,
	topic    string,
	handler func(context.Context, kafka.Message) error,
) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic: topic,
		}),
		handler: handler,
	}
}


func (c *Consumer) Start(ctx context.Context) {
	log.Println("🚀 Kafka consumer started")


	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Println("❌ Fetch error:", err)
			continue
		}
 
		if err := c.handler(ctx, msg); err != nil {
			log.Println("❌ Handler failed:", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Println("❌ Commit failed:", err)
		}
	}
}

