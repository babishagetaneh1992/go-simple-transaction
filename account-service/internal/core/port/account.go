package port

import (
	"account/internal/core/domain"
	"context"
)

type AccountRepository interface {
	Create(ctx context.Context, name string) (*domain.Account, error)
	GetByID(ctx context.Context, id int64) (*domain.Account, error)
	LockByID(ctx context.Context, id int64) (*domain.Account, error)
}
