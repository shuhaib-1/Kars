package models

import "gorm.io/gorm"

type Wallet struct {
	gorm.Model
	UserID      uint     `json:"user_id"`
	TotalAmount float64  `json:"total_amount"`
}

type WalletHistory struct {
	gorm.Model
	WalletID uint    `json:"wallet_id"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
}
