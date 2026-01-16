CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,

    account_id BIGINT NOT NULL
        REFERENCES accounts(id)
        ON DELETE CASCADE,

    type TEXT NOT NULL CHECK (
        type IN (
            'DEPOSIT',
            'WITHDRAW',
            'TRANSFER_IN',
            'TRANSFER_OUT'
        )
    ),

    amount BIGINT NOT NULL CHECK (amount > 0),
    note TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
