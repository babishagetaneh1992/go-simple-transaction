package transaction

import "context"

type TransactionService interface {
	Deposit(ctx context.Context,  idempotencyKey string, accountID int64,  amount int64, note string) error 
	Withdraw(ctx context.Context, idempotencyKey string, accountID int64, amount int64, note string) error
	Transfer(ctx context.Context, fromAccountID int64, toAccountID int64, amount int64, note string) error
	History(ctx context.Context, accountID int64) ([]Transaction, error)
	Balance( ctx context.Context, accountID int64) (int64, error)

}