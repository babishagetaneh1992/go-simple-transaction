package main

import (
	"log"
	"net/http"
	"os"
	"transaction/internal/account"
	"transaction/internal/infrastructure/database"
	"transaction/internal/transaction"

	httpinfra "transaction/internal/infrastructure/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		// try loading from parent directory
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found")
		}
	}
	db, err := database.NewPostgres(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	accountRepo := account.NewPostgresRepository(db)
	transactionRepo := transaction.NewPostgresRepo(db)

	transactionService := transaction.NewTransactionService(db, accountRepo, transactionRepo)

	accountHandler := account.NewAccountHandler(accountRepo)
	transactionHandler := transaction.NewTransactionHandler(transactionService)

	router := httpinfra.NewRouter(
		accountHandler.Routes(),
		transactionHandler.Routes(),
	)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
