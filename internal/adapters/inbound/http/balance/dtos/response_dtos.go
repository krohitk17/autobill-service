package BalanceDtos

type UserBalanceItemDto struct {
	OtherUserID   string `json:"other_user_id"`
	OtherUserName string `json:"other_user_name"`
	NetAmount     int64  `json:"net_amount"`
	Currency      string `json:"currency"`
}

type UserBalanceResponseDto struct {
	UserID   string               `json:"user_id"`
	Balances []UserBalanceItemDto `json:"balances"`
}

type GroupBalanceItemDto struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	NetAmount int64  `json:"net_amount"`
	Currency  string `json:"currency"`
}

type GroupBalanceResponseDto struct {
	GroupID   string                `json:"group_id"`
	GroupName string                `json:"group_name"`
	Balances  []GroupBalanceItemDto `json:"balances"`
}

type SimplifiedDebtDto struct {
	FromUserID   string `json:"from_user_id"`
	FromUserName string `json:"from_user_name"`
	ToUserID     string `json:"to_user_id"`
	ToUserName   string `json:"to_user_name"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
}

type SimplifiedDebtsResponseDto struct {
	GroupID string              `json:"group_id"`
	Debts   []SimplifiedDebtDto `json:"debts"`
}
