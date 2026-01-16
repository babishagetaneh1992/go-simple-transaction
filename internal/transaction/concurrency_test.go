package transaction_test

import (
	"context"
	"sync"
	"testing"
	"transaction/internal/account"
	"transaction/internal/transaction"
)

func TestConcurrentWithdraw(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
    key := "abc-123"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	acc, _ := accountRepo.Create(ctx, "Concurrent User")
	service.Deposit(ctx, key, acc.ID, 10_000, "initial")

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	withdraw := func() {
		defer wg.Done()
		err := service.Withdraw(ctx, acc.ID, 8_000, "race")
		errors <- err
	}

	wg.Add(2)
	go withdraw()
	go withdraw()
	wg.Wait()
	close(errors)

	success := 0
	failed := 0

	for err := range errors {
		if err == nil {
			success++
		} else {
			failed++
		}
	}

	if success != 1 || failed != 1 {
		t.Fatalf("expected 1 success and 1 failure, got %d success, %d failure", success, failed)
	}
}




func TestConcurrentTransfers(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	from, _ := accountRepo.Create(ctx, "From")
	to, _ := accountRepo.Create(ctx, "To")

	service.Deposit(ctx, key, from.ID, 50_000, "fund")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.Transfer(ctx, from.ID, to.ID, 10_000, "parallel")
		}()
	}

	wg.Wait()

	entries, _ := txRepo.ListByAccount(ctx, from.ID)

	if len(entries) > 6 { // 1 deposit + 5 transfers max
		t.Fatal("double-spend detected")
	}
}
