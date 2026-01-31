package transaction

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID     uuid.UUID
	AggregateType  string
	AggregateID    int64 
	EventType       string
	Payload         []byte
	Status          string
	CreatedAt       time.Time
	ProcessedAt     *time.Time 

}