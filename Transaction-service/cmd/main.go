package main

import (
	"context"
	"log"
	"net/http"
	"os"

	//"strings"
	//"transaction/internal/account"
	"transaction/internal/infrastructure/database"
	"transaction/internal/infrastructure/kafka"
	"transaction/internal/transaction"
	"transaction/pb"

	//"transaction/Transaction-service/internal/infrastructure/database"

	// "transaction/Transaction-service/internal/account"
	// "transaction/Transaction-service/internal/infrastructure/database"
	// "transaction/Transaction-service/internal/transaction"

	httpinfra "transaction/internal/infrastructure/http"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
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


	// grpc connection
	accountConn, err := grpc.Dial(
		os.Getenv("ACCOUNT_SERVICE_GRPC_ADDR"),
		grpc.WithInsecure(),
	)
	if  err != nil {
		log.Fatal("failed to connect to account service:", err)
	}

	defer accountConn.Close()

	accountClient := pb.NewAccountServiceClient(accountConn)


	//accountRepo := account.NewPostgresRepository(db)
	transactionRepo := transaction.NewPostgresRepo(db)

	
	//transactionHandler := transaction.NewTransactionHandler(transactionService)
	

	//accountHandler := account.NewAccountHandler(accountRepo, transactionHandler.Balance)

	outboxRepo := transaction.NewPostgresOutboxRepository(db)
	//publisher := transaction.NewLogPublisher()


	//services
	transactionService := transaction.NewTransactionService(
		db, 
		accountClient,
		transactionRepo,
		
	 )

	 transactionHandler := transaction.NewTransactionHandler(transactionService)


	producer := kafka.NewProducer([]string{"localhost:9092"})
	worker := transaction.NewWorker(
		outboxRepo,
		producer,
		"transaction.events",
	)



	router := httpinfra.NewRouter(
		//accountHandler.Routes(),
		transactionHandler.Routes(),
	)


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	log.Println("🚀 Transaction Service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))


	
}
