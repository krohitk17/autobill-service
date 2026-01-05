package SplitDtos

import "time"

type ParticipantResponseDto struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ShareAmount int64  `json:"share_amount"`
	Currency    string `json:"currency"`
	IsSettled   bool   `json:"is_settled"`
}

type SplitResponseDto struct {
	ID            string                   `json:"id"`
	Type          string                   `json:"type"`
	DivisionType  string                   `json:"division_type"`
	TotalAmount   int64                    `json:"total_amount"`
	Currency      string                   `json:"currency"`
	Description   string                   `json:"description"`
	GroupID       string                   `json:"group_id,omitempty"`
	CreatedByID   string                   `json:"created_by_id"`
	CreatedAt     time.Time                `json:"created_at"`
	IsFinalized   bool                     `json:"is_finalized"`
	SimplifyDebts *bool                    `json:"simplify_debts"`
	Participants  []ParticipantResponseDto `json:"participants"`
}

type SplitListResponseDto struct {
	Splits     []SplitResponseDto `json:"splits"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalItems int64              `json:"total_items"`
	TotalPages int                `json:"total_pages"`
}
