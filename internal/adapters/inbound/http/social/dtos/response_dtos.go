package SocialDtos

type RequestType string

const (
	Sent     RequestType = "sent"
	Received RequestType = "received"
)

type CreateFriendRequestResponseDto struct {
	RequestId string `json:"request_id"`
}

type GetFriendRequestsListResponseDto struct {
	Type       RequestType        `json:"type"`
	Users      []FriendRequestDto `json:"users"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalItems int64              `json:"total_items"`
	TotalPages int                `json:"total_pages"`
}

type FriendRequestDto struct {
	RequestId string       `json:"request_id"`
	Id        string       `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Status    FriendStatus `json:"status"`
}

type FriendStatus string

const (
	FriendPending  FriendStatus = "PENDING"
	FriendAccepted FriendStatus = "ACCEPTED"
	FriendRejected FriendStatus = "REJECTED"
)

type GetFriendsListResponseDto struct {
	Id         string    `json:"id"`
	Friends    []UserDto `json:"friends"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalItems int64     `json:"total_items"`
	TotalPages int       `json:"total_pages"`
}

type UserDto struct {
	Id     string       `json:"id"`
	Name   string       `json:"name"`
	Email  string       `json:"email"`
	Status FriendStatus `json:"status"`
}
