package transaction

import "context"

type OutboxRepository interface {
	Add(ctx context.Context, event *OutboxEvent) error 
	FetchPending(ctx context.Context,  limit int) ([]OutboxEvent, error )
	MarkProcessed(ctx context.Context, id string) error
}