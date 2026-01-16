package transaction_test

import (
	"context"
	"testing"
	"transaction/internal/account"
	"transaction/internal/transaction"
	// "transaction/Transaction-service/internal/account"
	// "transaction/Transaction-service/internal/transaction"
	//"transaction/internal/account"
	//"transaction/internal/transaction"
)

func TestDeposit_CreatesLedgerEntry(t *testing.T) {
      db := setupTestDB(t)
	  ctx := context.Background()
	  key := "abc-123"

	  accountRepo := account.NewPostgresRepository(db)
	  txRepo := transaction.NewPostgresRepo(db)
	  service := transaction.NewTransactionService(db, accountRepo, txRepo)

	  acc, err := accountRepo.Create(ctx, "Charlie")
	  if err != nil {
		  t.Fatalf("❌ Failed to create account: %v", err)
	  }

	  err = service.Deposit(ctx, key, acc.ID, 5_000, "salary")
      if err != nil {
		  t.Fatalf("❌ Deposit failed for account %s: %v", acc.Name, err)

	  }

	  entries, err := txRepo.ListByAccount(ctx, acc.ID)
	  if err != nil {
		  t.Fatalf("❌ Could not list transactions for account %s: %v", acc.Name, err)

	  }

	  if len(entries) != 1 {
		  t.Fatalf("❌ Expected 1 ledger entry, got %d for account %s", len(entries), acc.Name)

	  }

	  entry := entries[0]
	  
	  if entry.Type != transaction.TypeDeposit {
		t.Fatalf("❌ expected DEPOSIT, got %s", entry.Type)
	  }

	  if entry.Amount != 5_000 {
		t.Fatalf("expected amount 5000, got %d", entry.Amount)
	  }



}



func TestWithdraw_CreatesLedgerEntry(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"
	

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	acc, _ := accountRepo.Create(ctx, "Bob")
	service.Deposit(ctx, key, acc.ID, 10_000, "funding")

	err := service.Withdraw(ctx, acc.ID, 3_000, "rent")
	if err != nil {
		t.Fatal(err)
	}

	entries, _ := txRepo.ListByAccount(ctx, acc.ID)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Type != transaction.TypeWithdraw {
		t.Fatalf("expected WITHDRAW, got %s", entries[0].Type)
	}
}




func TestTransfer_CreatesTwoLedgerEntries(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	from, _ := accountRepo.Create(ctx, "Sender")
	to, _ := accountRepo.Create(ctx, "Receiver")

	service.Deposit(ctx, key, from.ID, 20_000, "funding")

	err := service.Transfer(ctx, from.ID, to.ID, 7_000, "payment")
	if err != nil {
		t.Fatal(err)
	}

	fromEntries, _ := txRepo.ListByAccount(ctx, from.ID)
	toEntries, _ := txRepo.ListByAccount(ctx, to.ID)

	if len(fromEntries) != 2 {
		t.Fatalf("expected 2 entries for sender, got %d", len(fromEntries))
	}

	if len(toEntries) != 1 {
		t.Fatalf("expected 1 entry for receiver, got %d", len(toEntries))
	}

	if fromEntries[0].Type != transaction.TypeTransferOut {
		t.Fatalf("expected TRANSFER_OUT, got %s", fromEntries[0].Type)
	}

	if toEntries[0].Type != transaction.TypeTransferIn {
		t.Fatalf("expected TRANSFER_IN, got %s", toEntries[0].Type)
	}
}
