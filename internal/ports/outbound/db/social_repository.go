package RepositoryPorts

import (
	"context"

	Domain "autobill-service/internal/domain"

	"github.com/google/uuid"
)

type FriendRequestType string

const (
	FriendRequestSent     FriendRequestType = "sent"
	FriendRequestReceived FriendRequestType = "received"
)

type SocialRepositoryPort interface {
	GetFriendRequestsList(ctx context.Context, userId uuid.UUID, requestType FriendRequestType, limit, offset int) ([]*Domain.FriendRequest, int64, error)
	CreateFriendRequest(ctx context.Context, senderId uuid.UUID, receiverId uuid.UUID) (*Domain.FriendRequest, error)
	AcceptFriendRequest(ctx context.Context, receiverId uuid.UUID, requestId uuid.UUID) error
	RejectFriendRequest(ctx context.Context, receiverId uuid.UUID, requestId uuid.UUID) error
	CheckExistingRequest(ctx context.Context, senderId, receiverId uuid.UUID) (bool, error)
	CheckFriendship(ctx context.Context, userId, friendId uuid.UUID) (bool, error)

	GetFriendsList(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*Domain.User, int64, error)
	RemoveFriend(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error
}
