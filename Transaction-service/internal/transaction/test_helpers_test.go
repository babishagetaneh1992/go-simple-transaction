package transaction_test

import (
	"context"
	"database/sql"
	"testing"
	"transaction/pb"

	"google.golang.org/grpc"
)

type MockAccountClient struct{}

func (m *MockAccountClient) GetAccount(ctx context.Context, in *pb.GetAccountRequest, opts ...grpc.CallOption) (*pb.GetAccountResponse, error) {
	return &pb.GetAccountResponse{
		Id:       in.AccountId,
		Name:     "Test Account",
		IsExists: true,
		IsActive: true,
	}, nil
}

type TestAccount struct {
	ID   int64
	Name string
}

func createTestAccount(t *testing.T, db *sql.DB, name string) *TestAccount {
	var id int64
	err := db.QueryRow("INSERT INTO accounts (name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}
	return &TestAccount{ID: id, Name: name}
}
