package transaction_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"transaction/internal/transaction"

	"github.com/joho/godotenv"
)

func setupTestDB(t *testing.T) *sql.DB {
	_ = godotenv.Load("../../.env")
	rawURL := os.Getenv("DATABASE_URL")
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid DATABASE_URL: %v", err)
	}

	// Force usage of transaction_test database for tests
	u.Path = "/transaction_test"

	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}

	db.Exec(`
CREATE TABLE IF NOT EXISTS accounts (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('DEPOSIT', 'WITHDRAW', 'TRANSFER_IN', 'TRANSFER_OUT')),
    amount BIGINT NOT NULL CHECK (amount > 0),
    note TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

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

	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, &MockAccountClient{}, txRepo)

	acc := createTestAccount(t, db, "Alice")
	var err error
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
	key := "abc-123"

	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, &MockAccountClient{}, txRepo)

	acc := createTestAccount(t, db, "Bob")
	var err error
	t.Logf("👤 Created account: %s", acc.Name)

	t.Run("Attempt Overdraft", func(t *testing.T) {
		amount := int64(5_000)
		t.Logf("💸 Attempting to withdraw %s from empty account...", formatMoney(amount))
		err = service.Withdraw(ctx, key, acc.ID, amount, "bad withdraw")
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

	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, &MockAccountClient{}, txRepo)

	from := createTestAccount(t, db, "Sender")
	to := createTestAccount(t, db, "Receiver")
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

	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, &MockAccountClient{}, txRepo)

	acc := createTestAccount(t, db, "Ledger User")
	t.Logf("👤 Created account: %s", acc.Name)

	t.Run("Perform Multiple Transactions", func(t *testing.T) {
		ops := []struct {
			name   string
			action func() error
		}{
			{"Deposit 100.00", func() error { return service.Deposit(ctx, key, acc.ID, 10_000, "funding") }},
			{"Withdraw 25.00", func() error { return service.Withdraw(ctx, key, acc.ID, 2_500, "expense") }},
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

	txRepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, &MockAccountClient{}, txRepo)

	acc := createTestAccount(t, db, "Idempotent User")

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
