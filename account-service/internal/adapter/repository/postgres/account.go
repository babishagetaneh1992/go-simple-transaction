package postgres

import (
	"account/internal/core/domain"
	"account/internal/infrastructure/database"
	"context"
	"database/sql"
	"errors"
)

type PostgresRepository struct {
	db database.DBTX
}

func NewPostgresRepository(db database.DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, name string) (*domain.Account, error) {
	query := `
         INSERT INTO accounts (name)
		 VALUES ($1)
		 RETURNING id, name, created_at, updated_at
         `

	acc := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&acc.ID, &acc.Name, &acc.CreatedAt, &acc.UpdatedAt)

	return acc, err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Account, error) {
	query := `
	         SELECT id, name, created_at, updated_at
	         FROM accounts
	         WHERE id = $1
	   `

	acc := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&acc.ID, &acc.Name, &acc.CreatedAt, &acc.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("account not found")
	}

	return acc, err
}

func (r *PostgresRepository) LockByID(ctx context.Context, id int64) (*domain.Account, error) {
	query := `SELECT id, name, created_at, updated_at
	         FROM accounts
			 WHERE id = $1
			 FOR UPDATE
	`

	acc := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&acc.ID, &acc.Name, &acc.CreatedAt, &acc.UpdatedAt)

	if err != nil {
		return nil, errors.New("account not found")
	}

	return acc, err
}
