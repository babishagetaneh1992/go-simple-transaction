package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type TransactionCreatedEvent struct {
	AccountID int64  `json:"account_id"`
	Amount    int64  `json:"amount"`
	Type      string `json:"type"`
}

type Notifier interface {
	Notify(ctx context.Context, msg string) error
}


func TransactionCreatedHandler(notifier Notifier,) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var event TransactionCreatedEvent

		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}

		message := "Transaction processed for account " +
		     fmt.Sprint(event.AccountID)

			 log.Println("📨 Sending notification:", message)

			 return notifier.Notify(ctx, message)
	}
}