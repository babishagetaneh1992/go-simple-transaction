package transaction

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"transaction/internal/account"

	"github.com/google/uuid"
	//"transaction/Transaction-service/internal/account"
	//"strings"
	//"runtime/debug"
	//"transaction/internal/account"
)

type Service struct {
	db              *sql.DB
	accountRepo     account.AccountRepository
	transactionRepo TransactionRepository
	idempotencyRepo IdempotencyRepository
}

func NewTransactionService(
	db *sql.DB,
	accountRepo account.AccountRepository,
	transactionRepo TransactionRepository,
) *Service {
	return &Service{
		db:              db,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		idempotencyRepo: NewPostgresIdempotencyRepo(db),
	}
}

func (s *Service) Deposit(ctx context.Context, idempotencyKey string, accountID int64, amount int64, note string) error {

	


	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	accountRepo := account.NewPostgresRepository(tx)
	transactionRepo := NewPostgresRepo(tx)
	idemRepo:= NewPostgresIdempotencyRepo(tx)
	outboxRepo := NewPostgresOutboxRepository(tx)



	// try insert first
	if idempotencyKey != ""{
	inserted, err := idemRepo.TryInsert(ctx, idempotencyKey, "deposit")
	if err != nil {
		return  err
	}

	if !inserted {
		return nil// already proccesed
	}
  }

	_, err = accountRepo.LockByID(ctx, accountID)
	if err != nil {
		return err
	}

	// newBalance := acc.Balance + amount

	// if err := accountRepo.UpdateBalance(ctx, acc.ID, newBalance); err != nil {
	// 	return err
	// }
    
	// balance, err := transactionRepo.BalanceByAccount(ctx, accountID)
	// if balance < amount {
	// 	return  errors.New("insufficient funds")
	// }


	// write ledger entry
	if err := transactionRepo.Create(ctx, &Transaction{
		AccountID: accountID,
		Amount:    amount,
		Type:      TypeDeposit,
		Note:      note,
	}); err != nil {
		return err
	
	}

	payload := map[string]interface{}{
	"account_id": accountID,
	"amount":     amount,
	"type":       "deposit",
	"note":       note,
}

payloadJSON, err := json.Marshal(payload)
if err != nil {
	return err
}


	event := OutboxEvent {
		ID:      uuid.New(),
		AggregateType : "account",
		AggregateID: accountID,
		EventType: "transaction.created",
		Payload: payloadJSON,
	}


	if err := outboxRepo.Add(ctx, &event); err != nil {
		return  err
	} 


	// save idempotency record
	if idempotencyKey != "" {
		resp, _ := json.Marshal(map[string] string{"status":"ok"})
		if err := idemRepo.Save(ctx, idempotencyKey, "deposit", resp); err != nil {
			return err
		}
	}

	return tx.Commit()

}



func (s *Service) Withdraw(ctx context.Context, accountID int64, amount int64, note string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	accountRepo := account.NewPostgresRepository(tx)
	transactionRepo := NewPostgresRepo(tx)

	acc, err := accountRepo.LockByID(ctx, accountID)
	if err != nil {
		return err
	}


	// newBalance := acc.Balance - amount

	// if err := accountRepo.UpdateBalance(ctx, acc.ID, newBalance); err != nil {
	// 	return err
	// }

	// compute balance inside Tx
    balance, err := transactionRepo.BalanceByAccount(ctx, accountID)
	if err != nil {
		return  err
	}

  
	if balance < amount {
		return errors.New("insufficient funds")
	}


	// write ledger entry
	if err := transactionRepo.Create(ctx, &Transaction{
		AccountID: acc.ID,
		Type:      TypeWithdraw,
		Amount:    amount,
		Note:      note,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) Transfer(ctx context.Context, fromAccountID int64, toAccountID int64, amount int64, note string) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if fromAccountID == toAccountID {
		return errors.New("cannot transfer to the same account")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	accountRepo := account.NewPostgresRepository(tx)
	transactionRepo := NewPostgresRepo(tx)

	// // Lock accounts in ID order (deadlock prevention)
	// first, second := fromAccountID, toAccountID
	// if first > second {
	// 	first, second = second, first
	// }

	fromAcc, err := accountRepo.LockByID(ctx, fromAccountID)
	if err != nil {
		return err
	}

	toAcc, err := accountRepo.LockByID(ctx, toAccountID)
	if err != nil {
		return err
	}

	// if fromAcc.Balance < amount {
	// 	return errors.New("insufficient funds")
	// }

	// check balance
	balance, err := transactionRepo.BalanceByAccount(ctx, fromAccountID)
	if err != nil {
		return  err
	}

	// if err := accountRepo.UpdateBalance(ctx, fromAcc.ID, fromAcc.Balance-amount); err != nil {
	// 	return err
	// }

	if balance < amount {
		return errors.New("Insufficient funds")
	}

	// if err := accountRepo.UpdateBalance(ctx, toAcc.ID, toAcc.Balance+amount); err != nil {
	// 	return err
	// }

	// debit
	if err := transactionRepo.Create(ctx, &Transaction{
		AccountID: fromAcc.ID,
		Type:      TypeTransferOut,
		Amount:    amount,
		Note:      fmt.Sprintf("To account %d: %s", toAcc.ID, note),
	}); err != nil {
		return err
	}

	//credit
	if err := transactionRepo.Create(ctx, &Transaction{
		AccountID: toAcc.ID,
		Type:      TypeTransferIn,
		Amount:    amount,
		Note:      fmt.Sprintf("From account %d: %s", fromAcc.ID, note),
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s * Service) History(ctx context.Context, accountID int64 )([]Transaction, error) {
	_, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.ListByAccount(ctx, accountID)
}


func (s *Service) Balance(ctx context.Context, accountID int64) (int64, error) {

	_, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return  0, err
	}
    
	return  s.transactionRepo.BalanceByAccount(ctx, accountID)
}





