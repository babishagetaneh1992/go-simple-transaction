package test

// import (
// 	"context"
// 	"database/sql"
// 	"os"
// 	"testing"
// 	"transaction/internal/account"
// 	"transaction/internal/transaction"

// 	"github.com/joho/godotenv"
// )


// func setupTestDB(t *testing.T) *sql.DB {
// 	  _ = godotenv.Load("../../.env")
//       db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
//       if err != nil {
//             t.Fatal(err)
//       }

// 	  t.Cleanup(func() {
// 		db.Exec("TRUNCATE accounts, transactions RESTART IDENTITY CASCADE")
// 		db.Close()
// 	})

//       return db
// }

