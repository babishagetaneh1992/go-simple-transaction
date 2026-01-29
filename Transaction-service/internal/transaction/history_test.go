package transaction_test

import (
	"context"
	"testing"
	"transaction/internal/transaction"
)

func TestTransactionHistory(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"

	txrepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, &MockAccountClient{}, txrepo)

	acc := createTestAccount(t, db, "History User")
	service.Deposit(ctx, key, acc.ID, 3_000, "income")
	service.Withdraw(ctx, key, acc.ID, 1_000, "expenses")

	history, err := service.History(ctx, acc.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(history) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(history))
	}

}
