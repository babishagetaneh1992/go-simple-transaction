package transaction

import "time"

type Type string


const (
	TypeDeposit     Type = "DEPOSIT"
	TypeWithdraw    Type = "WITHDRAW"
	TypeTransferIn  Type = "TRANSFER_IN"
	TypeTransferOut Type = "TRANSFER_OUT"
)



type Transaction struct {
	ID        int64     `json:"id"`
	AccountID int64     `json:"account_id"`
	Type      Type      `json:"type"`
	Amount    int64     `json:"amount"` // cents, always positive
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}