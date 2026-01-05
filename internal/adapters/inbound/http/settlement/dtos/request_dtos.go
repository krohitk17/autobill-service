package SettlementDtos

type CreateSettlementRequestDto struct {
	SplitID        string `json:"split_id" validate:"required"`
	PayeeID        string `json:"payee_id" validate:"required"`
	Amount         int64  `json:"amount" validate:"required,gt=0"`
	Currency       string `json:"currency" validate:"required,oneof=INR USD EUR"`
	IdempotencyKey string `json:"idempotency_key" validate:"omitempty,max=64"`
}
