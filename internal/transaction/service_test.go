package transaction_test

import (
	"context"
	"database/sql"
	"fmt"
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

	db.Exec(`
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key TEXT PRIMARY KEY,
    operation TEXT NOT NULL,
    response JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
`)

	t.Cleanup(func() {
		db.Exec("TRUNCATE accounts, transactions, idempotency_keys RESTART IDENTITY CASCADE")
		db.Close()
	})

	return db
}

func formatMoney(cents int64) string {
	return fmt.Sprintf("%.2f", float64(cents)/100.0)
}

func assertBalance(t *testing.T, s transaction.TransactionService, accountID int64, expected int64) {
	t.Helper()
	balance, err := s.Balance(context.Background(), accountID)
	if err != nil {
		t.Fatalf("❌ Failed to get balance: %v", err)
	}
	if balance != expected {
		t.Errorf("❌ Balance mismatch: expected %s, got %s", formatMoney(expected), formatMoney(balance))
	} else {
		t.Logf("✅ Balance verified: %s", formatMoney(balance))
	}
}

func TestDeposit_IncreasesBalance(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	acc, err := accountRepo.Create(ctx, "Alice")
	if err != nil {
		t.Fatalf("❌ Failed to create account: %v", err)
	}
	t.Logf("👤 Created account: %s", acc.Name)

	t.Run("Perform Deposit", func(t *testing.T) {
		amount := int64(10_000)
		t.Logf("💵 Depositing %s...", formatMoney(amount))
		err = service.Deposit(ctx, key, acc.ID, amount, "initial deposit")
		if err != nil {
			t.Fatalf("❌ Deposit failed: %v", err)
		}
	})

	t.Run("Check Balance", func(t *testing.T) {
		assertBalance(t, service, acc.ID, 10_000)
	})
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
	t.Logf("👤 Created account: %s", acc.Name)

	t.Run("Attempt Overdraft", func(t *testing.T) {
		amount := int64(5_000)
		t.Logf("💸 Attempting to withdraw %s from empty account...", formatMoney(amount))
		err = service.Withdraw(ctx, acc.ID, amount, "bad withdraw")
		if err == nil {
			t.Fatalf("❌ Withdraw succeeded but should have failed due to insufficient funds")
		}
		t.Log("✅ Withdraw failed as expected (Insufficient Funds)")
	})
}

func TestTransfer_Atomicity(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	from, _ := accountRepo.Create(ctx, "Sender")
	to, _ := accountRepo.Create(ctx, "Receiver")
	t.Logf("👤 Created accounts: %s -> %s", from.Name, to.Name)

	t.Run("Setup Initial Funds", func(t *testing.T) {
		err := service.Deposit(ctx, key, from.ID, 20_000, "funding")
		if err != nil {
			t.Fatalf("❌ Setup failed: %v", err)
		}
		assertBalance(t, service, from.ID, 20_000)
	})

	t.Run("Execute Transfer", func(t *testing.T) {
		amount := int64(15_000)
		t.Logf("🔄 Transferring %s from %s to %s...", formatMoney(amount), from.Name, to.Name)
		err := service.Transfer(ctx, from.ID, to.ID, amount, "payment")
		if err != nil {
			t.Fatalf("❌ Transfer failed: %v", err)
		}
	})

	t.Run("Verify Final Balances", func(t *testing.T) {
		assertBalance(t, service, from.ID, 5_000)
		assertBalance(t, service, to.ID, 15_000)
	})
}

func TestLedgerDerivedBalance(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-123"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	acc, _ := accountRepo.Create(ctx, "Ledger User")
	t.Logf("👤 Created account: %s", acc.Name)

	t.Run("Perform Multiple Transactions", func(t *testing.T) {
		ops := []struct {
			name   string
			action func() error
		}{
			{"Deposit 100.00", func() error { return service.Deposit(ctx, key, acc.ID, 10_000, "funding") }},
			{"Withdraw 25.00", func() error { return service.Withdraw(ctx, acc.ID, 2_500, "expense") }},
			{"Deposit 10.00", func() error { return service.Deposit(ctx, key+"-2", acc.ID, 1_000, "refund") }},
		}

		for _, op := range ops {
			t.Logf("▶️ %s", op.name)
			if err := op.action(); err != nil {
				t.Fatalf("❌ %s failed: %v", op.name, err)
			}
		}
	})

	t.Run("Verify Derived Balance", func(t *testing.T) {
		// 100 - 25 + 10 = 85
		assertBalance(t, service, acc.ID, 8_500)
	})
}

func TestDeposit_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	key := "abc-idempotent"

	accountRepo := account.NewPostgresRepository(db)
	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txRepo)

	acc, _ := accountRepo.Create(ctx, "Idempotent User")

	t.Run("First Deposit", func(t *testing.T) {
		t.Log("1️⃣ Performing first deposit...")
		err := service.Deposit(ctx, key, acc.ID, 10_000, "once")
		if err != nil {
			t.Fatalf("❌ First deposit failed: %v", err)
		}
		assertBalance(t, service, acc.ID, 10_000)
	})

	t.Run("Second Deposit (Duplicate Key)", func(t *testing.T) {
		t.Log("2️⃣ Performing second deposit with same key...")
		err := service.Deposit(ctx, key, acc.ID, 10_000, "twice")
		if err != nil {
			t.Fatalf("❌ Second deposit returned error: %v", err)
		}
		// Balance should NOT increase
		assertBalance(t, service, acc.ID, 10_000)
		t.Log("✅ Balance correctly remained unchanged")
	})
}
