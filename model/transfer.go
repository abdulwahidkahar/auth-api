package model

type Transfer struct {
	ID           int    `json:"id"`
	FromWalletID int    `json:"from_wallet_id"`
	ToWalletID   int    `json:"to_wallet_id"`
	Amount       int64  `json:"amount"`
	CreatedAt    string `json:"created_at"`
}

type TransferRequest struct {
	FromWalletID int   `json:"from_wallet_id"`
	ToWalletID   int   `json:"to_wallet_id"`
	Amount       int64 `json:"amount"`
}

type TransferResponse struct {
	ID           int    `json:"id"`
	FromWalletID int    `json:"from_wallet_id"`
	ToWalletID   int    `json:"to_wallet_id"`
	Amount       int64  `json:"amount"`
	CreatedAt    string `json:"created_at"`
}
