package BalanceApplicationDtos

type UserBalanceItemResult struct {
	OtherUserID   string
	OtherUserName string
	NetAmount     int64
	Currency      string
}

type UserBalanceResult struct {
	UserID   string
	Balances []UserBalanceItemResult
}

type GroupBalanceItemResult struct {
	UserID    string
	UserName  string
	NetAmount int64
	Currency  string
}

type GroupBalanceResult struct {
	GroupID   string
	GroupName string
	Balances  []GroupBalanceItemResult
}

type SimplifiedDebtResult struct {
	FromUserID   string
	FromUserName string
	ToUserID     string
	ToUserName   string
	Amount       int64
	Currency     string
}

type SimplifiedDebtsResult struct {
	GroupID string
	Debts   []SimplifiedDebtResult
}
