package GroupDtos

type CreateGroupRequestDto struct {
	Name          string `json:"name" validate:"required"`
	SimplifyDebts *bool  `json:"simplify_debts"`
}

type UpdateGroupRequestDto struct {
	Name          *string `json:"name"`
	SimplifyDebts *bool   `json:"simplify_debts"`
}

type AddMemberRequestDto struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=ADMIN MEMBER"`
}

type UpdateMemberRoleRequestDto struct {
	Role string `json:"role" validate:"required,oneof=OWNER ADMIN MEMBER"`
}
