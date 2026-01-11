package transaction

import "context"

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	ListByAccount(ctx context.Context, accountID int64) ([]Transaction, error)
}