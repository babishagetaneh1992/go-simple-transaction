package transaction

import (
	"context"
	"fmt"
	"log"
	"time"
)

type OutboxWorker struct {
	outboxRepo OutboxRepository
	publisher  EventPublisher
	interval   time.Duration
	batchSize  int
}

type Worker struct {
	repo      OutboxRepository
	publisher Publisher
	topic     string
}

func NewOutboxWorker(outboxRepo OutboxRepository, publisher EventPublisher) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		interval:   1 * time.Second,
		batchSize:  10,
	}
}

func NewWorker(repo OutboxRepository, publisher Publisher, topic string) *Worker {
	return &Worker{
		repo:      repo,
		publisher: publisher,
		topic:     topic,
	}
}

// func (w *Worker) processBatch(ctx context.Context) {
// 	events, err := w.repo.FetchPending(ctx, 10)
// 	if err != nil {
// 		log.Println("❌ Fetch pending failed:", err)
// 		return
// 	}

// 	for _, event := range events {
// 		if err := w.publisher.Publish(ctx, w.topic, w.publisher); err != nil {
// 			log.Println("❌ Publish failed:", err)
// 			continue // retry later
// 		}

// 		if err := w.repo.MarkProcessed(ctx, event.ID.String()); err != nil {
// 			log.Println("❌ Mark processed failed:", err)
// 		}
// 	}
// }

func (w *Worker) Start(ctx context.Context) {
	log.Println("🚀 Outbox worker started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Outbox worker stopped")
			return

		case <-ticker.C:
			events, err := w.repo.FetchPending(ctx, 10)
			if err != nil {
				log.Println("❌ Fetch pending failed:", err)
				return
			}

			for _, e := range events {
				if err := w.publisher.Publish(
					ctx,
					w.topic,
					fmt.Sprintf("%d", e.AggregateID),
					e.Payload,
				); err != nil {
					log.Println("❌ Publish failed:", err)
					continue
				}

				if err := w.repo.MarkProcessed(ctx, e.ID.String()); err != nil {
					log.Println("❌ Mark processed failed:", err)
				}
			}
		}
	}
}
