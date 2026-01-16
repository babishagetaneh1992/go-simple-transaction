package transaction

import (
	"context"
	"database/sql"
	"transaction/internal/infrastructure/database"
	//"encoding/json"
)

type IdempotencyRepository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Save(ctx context.Context, key, operation string, response []byte) error
	TryInsert(ctx context.Context, idempotencyKey string, operation string) (bool, error)
}

type PostgresIdempotencyRepo struct {
	db database.DBTX
}

func NewPostgresIdempotencyRepo(db database.DBTX) *PostgresIdempotencyRepo {
	return &PostgresIdempotencyRepo{db: db}
}

func (r *PostgresIdempotencyRepo) Get(ctx context.Context, key string) ([]byte, error) {
	var response []byte
	err := r.db.QueryRowContext(
		ctx,
		`SELECT response FROM idempotency_keys WHERE key = $1`,
		key,
	).Scan(&response)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return response, err
}

func (r *PostgresIdempotencyRepo) Save(ctx context.Context, key string, operation string, response []byte) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO idempotency_keys (key, operation, response)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE 
		 SET response = EXCLUDED.response, operation = EXCLUDED.operation`,
		key, operation, response,
	)
	return err
}

func (r *PostgresIdempotencyRepo) TryInsert(ctx context.Context, key string, operation string) (bool, error) {

	query := `
	       INSERT INTO idempotency_keys (key, operation, response)
		   VALUES ($1, $2, '{}' ::jsonb)
		   ON CONFLICT (key) DO NOTHING
		   RETURNING key
	`

	var insertedKey string
	err := r.db.QueryRowContext(ctx, query, key, operation).Scan(&insertedKey)

	if err == sql.ErrNoRows {
		return false, nil // already exist
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
