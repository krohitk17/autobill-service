package SettlementApplicationDtos

import "time"

type CreateSettlementInput struct {
	SplitID        string
	PayeeID        string
	Amount         int64
	Currency       string
	IdempotencyKey string
}

type SettlementResult struct {
	ID        string
	SplitID   string
	PayerID   string
	PayerName string
	PayeeID   string
	PayeeName string
	Amount    int64
	Currency  string
	Date      time.Time
	Confirmed bool
}

type SettlementListResult struct {
	Settlements []SettlementResult
	Page        int
	PageSize    int
	TotalItems  int64
}
