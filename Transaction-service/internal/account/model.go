package account

import "time"

type Account struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	//Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}