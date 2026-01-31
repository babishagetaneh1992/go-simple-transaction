package transaction

import (
	"context"
	"time"
	"transaction/internal/infrastructure/database"

	"github.com/google/uuid"
)

type PostgresOutboxRepository struct {
	db database.DBTX
}


func NewPostgresOutboxRepository (db database.DBTX) *PostgresOutboxRepository {
    return &PostgresOutboxRepository{db : db}
}



func ( r *PostgresOutboxRepository) Add(ctx context.Context, e *OutboxEvent) error {
	query := `
	       INSERT INTO outbox_events
           (id, aggregate_type, aggregate_id, event_type, payload)
		   VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		e.ID,
		e.AggregateType,
		e.AggregateID,
		e.EventType,
		e.Payload,

	)

	return  err
}



func (r *PostgresOutboxRepository) FetchPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := `
	       SELECT id, aggregate_id, aggregate_type, event_type, payload, status
		   FROM outbox_events
		   WHERE status = 'pending'
		   ORDER BY created_at
		   LIMIT $1
		  
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		err := rows.Scan(
			&e.ID,
			&e.AggregateID,
			&e.AggregateType,
			&e.EventType,
			&e.Payload,
			&e.Status,
			
			//&e.CreatedAt,
		)

		if err != nil {
			return  nil, err
		}

		events = append(events, e)
	}

	return  events, nil

	
}



func (r *PostgresOutboxRepository) MarkProcessed(ctx context.Context, id string ) error {
	query := `
	      UPDATE outbox_events
		  SET status = 'processed',
		    processed_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), uuid.MustParse(id))
	if err != nil {
		return  err
	}
	return  err
}

