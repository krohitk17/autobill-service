package SplitApplicationDtos

import "time"

type ParticipantInput struct {
	UserID      string
	ShareAmount int64
}

type CreateSplitInput struct {
	Type           string
	DivisionType   string
	TotalAmount    int64
	Currency       string
	Description    string
	GroupID        string
	SimplifyDebts  *bool
	IdempotencyKey string
	Participants   []ParticipantInput
}

type ParticipantResult struct {
	UserID      string
	UserName    string
	ShareAmount int64
	Currency    string
	IsSettled   bool
}

type SplitResult struct {
	ID            string
	Type          string
	DivisionType  string
	TotalAmount   int64
	Currency      string
	Description   string
	GroupID       string
	CreatedByID   string
	CreatedAt     time.Time
	SimplifyDebts *bool
	Participants  []ParticipantResult
}

type SplitListResult struct {
	Splits     []SplitResult
	Page       int
	PageSize   int
	TotalItems int64
}
