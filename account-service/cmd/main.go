package main

import (
	"log"
	"net"
	"net/http"
	"os"

	accountHttp "account/internal/adapter/handler/http"
	"account/internal/adapter/repository/postgres"
	accountGrpc "account/internal/grpc"
	"account/internal/infrastructure/database"
	httpInfra "account/internal/infrastructure/http"
	"account/pb"

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

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := database.NewPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Run migration
	migrationFile := "migration/create_account.sql"
	if _, err := os.Stat(migrationFile); os.IsNotExist(err) {
		// try checking relative to cmd if running from there
		migrationFile = "../migration/create_account.sql"
	}

	if schema, err := os.ReadFile(migrationFile); err != nil {
		log.Printf("⚠️ Could not read migration file %s: %v", migrationFile, err)
	} else {
		if _, err := db.Exec(string(schema)); err != nil {
			log.Printf("⚠️ Migration execution failed: %v", err)
		} else {
			log.Println("✅ Database migration applied successfully")
		}
	}

	repo := postgres.NewPostgresRepository(db)

	// Initialize HTTP Handler (keeping existing logic as requested)
	handler := accountHttp.NewAccountHandler(repo, nil)

	router := httpInfra.NewRouter(handler.Routes())

	// gRPC Server
	grpcServer := grpc.NewServer()
	accountGrpcHandler := accountGrpc.NewGrpcAccountServer(repo)
	pb.RegisterAccountServiceServer(grpcServer, accountGrpcHandler)

	go func() {
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Failed to listen on port 50051: %v", err)
		}
		log.Println("gRPC Server running on :50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	log.Println("HTTP Server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}


