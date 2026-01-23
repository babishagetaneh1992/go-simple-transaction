package main

import (
	"context"
	"log"
	"net/http"
	"os"
	//"strings"
	"transaction/internal/account"
	"transaction/internal/infrastructure/database"
	"transaction/internal/infrastructure/kafka"
	"transaction/internal/transaction"

	//"transaction/Transaction-service/internal/infrastructure/database"

	// "transaction/Transaction-service/internal/account"
	// "transaction/Transaction-service/internal/infrastructure/database"
	// "transaction/Transaction-service/internal/transaction"

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
	//transactionHandler := transaction.NewTransactionHandler(transactionService)
	transactionHandler := transaction.NewTransactionHandler(transactionService)

	accountHandler := account.NewAccountHandler(accountRepo, transactionHandler.Balance)

	outboxRepo := transaction.NewPostgresOutboxRepository(db)
	//publisher := transaction.NewLogPublisher()
	producer := kafka.NewProducer([]string{"localhost:9092"})
	worker := transaction.NewWorker(
		outboxRepo,
		producer,
		"transaction.events",
	)



	router := httpinfra.NewRouter(
		accountHandler.Routes(),
		transactionHandler.Routes(),
	)


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))


	
}
