package transaction

import (
	"context"
	"transaction/internal/infrastructure/database"
	//"transaction/Transaction-service/internal/infrastructure/database"
	//"transaction/internal/infrastructure/database"
	//"database/sql"
	//"transaction/internal/infrastructure/database"
)

var _ TransactionRepository = (*PostgresRepo)(nil)

type PostgresRepo struct {
	db database.DBTX
}

func NewPostgresRepo(db database.DBTX) *PostgresRepo {
	return &PostgresRepo{db: db}
}


func (r *PostgresRepo) Create(ctx context.Context, tx *Transaction) error {
	query := `
	        INSERT INTO transactions (account_id, amount, type, note) 
			VALUES ($1, $2, $3, $4)

	`
	_, err := r.db.ExecContext(ctx, query, tx.AccountID, tx.Amount, tx.Type, tx.Note)

	return err
}



func (r *PostgresRepo) ListByAccount(ctx context.Context, accountID int64) ([]Transaction, error) {
	query := `
		SELECT id, account_id, type, amount, note, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Transaction

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID,
			&t.AccountID,
			&t.Type,
			&t.Amount,
			&t.Note,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, t)
	}

	return result, nil
}


func (r *PostgresRepo) BalanceByAccount(ctx context.Context, accountId int64) (int64, error) {
	var balance int64
	err := r.db.QueryRowContext(ctx,
	 `
	     SELECT COALESCE(SUM(
		 CASE 
		     WHEN type IN ($2, $3) THEN amount
			 WHEN type IN ($4, $5) THEN -amount
			 ELSE 0
		 END
		 ), 0)
		 FROM transactions
		 WHERE account_id = $1
	`,
	     
	    accountId,
		TypeDeposit,
		TypeTransferIn,
		TypeWithdraw,
		TypeTransferOut,

	).Scan(&balance)
	
	return  balance, err

}