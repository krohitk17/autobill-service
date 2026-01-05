package SplitDtos

type ParticipantInput struct {
	UserID      string `json:"user_id" validate:"required"`
	ShareAmount int64  `json:"share_amount"`
}

type CreateSplitRequestDto struct {
	Type          string             `json:"type" validate:"required,oneof=GROUP DIRECT"`
	DivisionType  string             `json:"division_type" validate:"required,oneof=EQUAL CUSTOM"`
	TotalAmount   int64              `json:"total_amount" validate:"required,gt=0"`
	Currency      string             `json:"currency" validate:"required,oneof=INR USD EUR"`
	Description   string             `json:"description"`
	GroupID       string             `json:"group_id"`
	SimplifyDebts *bool              `json:"simplify_debts"`
	Participants  []ParticipantInput `json:"participants" validate:"required,min=1"`
}

type AddParticipantRequestDto struct {
	UserID      string `json:"user_id" validate:"required"`
	ShareAmount int64  `json:"share_amount"`
}

type UpdateParticipantRequestDto struct {
	ShareAmount int64 `json:"share_amount" validate:"required,gt=0"`
	IsSettled   *bool `json:"is_settled"`
}
