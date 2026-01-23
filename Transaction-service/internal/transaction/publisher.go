package transaction

import "context"

type EventPublisher interface {
	Publisher(ctx context.Context, event OutboxEvent) error 
}

type Publisher interface {
    Publish(ctx context.Context, topic string, key string, payload []byte) error 
}