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

    updated, err := accountRepo.GetByID(ctx, acc.ID)
    if err != nil {
        t.Fatalf("❌ Could not retrieve updated account %s: %v", acc.Name, err)
    }

    if updated.Balance != 10_000 {
        t.Fatalf("❌ Balance mismatch for %s: expected 10000, got %d", acc.Name, updated.Balance)
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

    // Ensure starting balance is zero
    err = accountRepo.UpdateBalance(ctx, acc.ID, 0)
    if err != nil {
        t.Fatalf("❌ Could not set balance for %s: %v", acc.Name, err)
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

    fromAcc, _ := accountRepo.GetByID(ctx, from.ID)
    toAcc, _ := accountRepo.GetByID(ctx, to.ID)

    if fromAcc.Balance != 5_000 {
        t.Fatalf("❌ Sender balance mismatch: expected 5000, got %d", fromAcc.Balance)
    }

    if toAcc.Balance != 15_000 {
        t.Fatalf("❌ Receiver balance mismatch: expected 15000, got %d", toAcc.Balance)
    }

    t.Logf("✅ Atomic transfer test passed: %s → %s", from.Name, to.Name)
}
