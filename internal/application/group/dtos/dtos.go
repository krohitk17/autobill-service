package GroupApplicationDtos

import "time"

type CreateGroupInput struct {
	Name          string
	SimplifyDebts *bool
}

type UpdateGroupInput struct {
	Name          *string
	SimplifyDebts *bool
}

type GroupResult struct {
	ID            string
	Name          string
	SimplifyDebts bool
	CreatedAt     time.Time
}

type GroupDetailResult struct {
	ID            string
	Name          string
	SimplifyDebts bool
	CreatedAt     time.Time
	Members       []MemberResult
}

type MemberResult struct {
	UserID string
	Name   string
	Email  string
	Role   string
}

type GroupListResult struct {
	Groups     []GroupResult
	Page       int
	PageSize   int
	TotalItems int64
}

type AddMemberInput struct {
	UserID string
	Role   string
}

type UpdateMemberRoleInput struct {
	Role string
}
