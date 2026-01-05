package SocialApplicationDtos

import (
	Domain "autobill-service/internal/domain"
)

type RequestType string

const (
	RequestTypeSent     RequestType = "sent"
	RequestTypeReceived RequestType = "received"
)

type FriendRequestResult struct {
	RequestId string
	UserId    string
	Name      string
	Email     string
	Status    Domain.FriendStatus
}

type FriendRequestListResult struct {
	Type       RequestType
	Requests   []FriendRequestResult
	Page       int
	PageSize   int
	TotalItems int64
}

type FriendResult struct {
	ID     string
	Name   string
	Email  string
	Status Domain.FriendStatus
}

type FriendsListResult struct {
	UserID     string
	Friends    []FriendResult
	Page       int
	PageSize   int
	TotalItems int64
}
