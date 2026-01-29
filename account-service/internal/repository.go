package account

import (
	"account/internal/core/domain"
	"context"
)

type AccountRepository interface {
	Create(ctx context.Context, name string) (*domain.Account, error)
	GetByID(ctx context.Context, id int64) (*domain.Account, error)

	//used only inside DB transactions
	//UpdateBalance(ctx context.Context, id int64, newBalance int64) error
	LockByID(ctx context.Context, id int64) (*domain.Account, error)
}