package test

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


func TestTransactionHistory(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	accountRepo := account.NewPostgresRepository(db)
	txrepo := transaction.NewPostgresRepo(db)
	service := transaction.NewTransactionService(db, accountRepo, txrepo)

	acc, _ := accountRepo.Create(ctx, "History User")
	service.Deposit(ctx, acc.ID, 3_000, "income")
	service.Withdraw(ctx, acc.ID, 1_000, "expenses")

	history, err := service.History(ctx, acc.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(history) != 2 {
       t.Fatalf("expected 2 entries, got %d", len(history))
	}

	
}
