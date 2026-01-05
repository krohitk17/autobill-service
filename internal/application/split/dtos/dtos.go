package SplitApplicationDtos

import "time"

type ParticipantInput struct {
	UserID      string
	ShareAmount int64
}

type CreateSplitInput struct {
	Type          string
	DivisionType  string
	TotalAmount   int64
	Currency      string
	Description   string
	GroupID       string
	SimplifyDebts *bool
	Participants  []ParticipantInput
}

type AddParticipantInput struct {
	UserID      string
	ShareAmount int64
}

type UpdateParticipantInput struct {
	ShareAmount int64
	IsSettled   *bool
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
	IsFinalized   bool
	SimplifyDebts *bool
	Participants  []ParticipantResult
}

type SplitListResult struct {
	Splits     []SplitResult
	Page       int
	PageSize   int
	TotalItems int64
}
