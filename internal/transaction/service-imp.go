package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	//"runtime/debug"
	"transaction/internal/account"
)

type Service struct {
	db              *sql.DB
	accountRepo     account.AccountRepository
	transactionRepo TransactionRepository
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
	}
}

func (s *Service) Deposit(ctx context.Context, accountID int64, amount int64, note string) error {
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





