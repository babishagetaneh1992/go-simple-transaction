package kafka

//import "transaction/internal/infrastructure/kafka"
import (
	"context"

	"github.com/segmentio/kafka-go"
)
type Producer struct {
	writer *kafka.Writer
} 

func NewProducer(brokers []string) *Producer {
	return  &Producer{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}


func (p *Producer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	return  p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key: []byte(key),
		Value: payload,
	})
}