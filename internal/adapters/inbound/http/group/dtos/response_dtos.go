package GroupDtos

import "time"

type GroupResponseDto struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	SimplifyDebts bool      `json:"simplify_debts"`
	CreatedAt     time.Time `json:"created_at"`
}

type GroupDetailResponseDto struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	SimplifyDebts bool                `json:"simplify_debts"`
	CreatedAt     time.Time           `json:"created_at"`
	Members       []MemberResponseDto `json:"members"`
}

type MemberResponseDto struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type GroupListResponseDto struct {
	Groups     []GroupResponseDto `json:"groups"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalItems int64              `json:"total_items"`
	TotalPages int                `json:"total_pages"`
}
