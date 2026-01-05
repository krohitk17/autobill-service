package SettlementDtos

import "time"

type SettlementResponseDto struct {
	ID        string    `json:"id"`
	SplitID   string    `json:"split_id"`
	PayerID   string    `json:"payer_id"`
	PayerName string    `json:"payer_name"`
	PayeeID   string    `json:"payee_id"`
	PayeeName string    `json:"payee_name"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Date      time.Time `json:"date"`
	Confirmed bool      `json:"confirmed"`
}

type SettlementListResponseDto struct {
	Settlements []SettlementResponseDto `json:"settlements"`
	Page        int                     `json:"page"`
	PageSize    int                     `json:"page_size"`
	TotalItems  int64                   `json:"total_items"`
	TotalPages  int                     `json:"total_pages"`
}
