package transaction

import (
	"context"
	"log"
)

type LogPublisher struct{}

func NewLogPublisher() *LogPublisher {
	return &LogPublisher{}
}

func (p *LogPublisher) Publisher(ctx context.Context, e OutboxEvent) error {
	log.Printf(
		"📤 EVENT %s | %s | aggregate=%s:%d | payload=%s",
		e.ID,
		e.EventType,
		e.AggregateType,
		e.AggregateID,
		
		string(e.Payload),
	)

	return  nil
}
