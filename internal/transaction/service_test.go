package transaction_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"transaction/internal/account"
	"transaction/internal/transaction"
	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) *sql.DB {
	  _ = godotenv.Load("../../.env")
      db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
      if err != nil {
            t.Fatal(err)
      }

	  t.Cleanup(func() {
		db.Exec("TRUNCATE accounts, transactions RESTART IDENTITY CASCADE")
		db.Close()
	})

      return db
}



func TestDeposit_IncreasesBalance(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()

    accountRepo := account.NewPostgresRepository(db)
    txRepo := transaction.NewPostgresRepo(db)
    service := transaction.NewTransactionService(db, accountRepo, txRepo)

    acc, err := accountRepo.Create(ctx, "Alice")
    if err != nil {
        t.Fatalf("❌ Failed to create account: %v", err)
    }

    err = service.Deposit(ctx, acc.ID, 10_000, "initial deposit")
    if err != nil {
        t.Fatalf("❌ Deposit failed for account %s: %v", acc.Name, err)
    }

  balance, err := service.Balance(ctx, acc.ID)
  if err != nil {
    t.Fatal(err)
  }

  if balance != 10_000 {
    t.Fatalf("expected balance 10000, got %d", balance)
  }

    t.Logf("✅ Deposit test passed for %s", acc.Name)
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()

    accountRepo := account.NewPostgresRepository(db)
    txRepo := transaction.NewPostgresRepo(db)
    service := transaction.NewTransactionService(db, accountRepo, txRepo)

    acc, err := accountRepo.Create(ctx, "Bob")
    if err != nil {
        t.Fatalf("❌ Failed to create account: %v", err)
    }

    
     

    // Attempt withdrawal
    err = service.Withdraw(ctx, acc.ID, 5_000, "bad withdraw")
    if err == nil {
        t.Fatalf("❌ Withdraw succeeded but should have failed due to insufficient funds for %s", acc.Name)
    }

    t.Logf("✅ Insufficient funds check passed for %s", acc.Name)
}

func TestTransfer_Atomicity(t *testing.T) {
    db := setupTestDB(t)
    ctx := context.Background()

    accountRepo := account.NewPostgresRepository(db)
    txRepo := transaction.NewPostgresRepo(db)
    service := transaction.NewTransactionService(db, accountRepo, txRepo)

    from, _ := accountRepo.Create(ctx, "Sender")
    to, _ := accountRepo.Create(ctx, "Receiver")

    err := service.Deposit(ctx, from.ID, 20_000, "funding")
    if err != nil {
        t.Fatalf("❌ Deposit to %s failed: %v", from.Name, err)
    }

    err = service.Transfer(ctx, from.ID, to.ID, 15_000, "payment")
    if err != nil {
        t.Fatalf("❌ Transfer from %s to %s failed: %v", from.Name, to.Name, err)
    }

    fromAcc, _ := service.Balance(ctx, from.ID)
    toAcc, _ := service.Balance(ctx, to.ID)

    if fromAcc != 5_000 {
        t.Fatalf("expected sender balance 5000, got %d", fromAcc)

    }

    if toAcc != 15_000 {
        t.Fatalf("expected reciever balance 15000, got %d", toAcc)
    }

  

    t.Logf("✅ Atomic transfer test passed: %s → %s", from.Name, to.Name)
}




func TestLedgerDerivedBalance(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	acc, _ := accountRepo.Create(ctx, "Ledger User")

	service.Deposit(ctx, acc.ID, 10_000, "funding")
	service.Withdraw(ctx, acc.ID, 2_500, "expense")
	service.Deposit(ctx, acc.ID, 1_000, "refund")

	balance, err := service.Balance(ctx, acc.ID)
	if err != nil {
		t.Fatal(err)
	}

	if balance != 8_500 {
		t.Fatalf("expected balance 8500, got %d", balance)
	}
}
