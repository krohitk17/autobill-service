package HttpPorts

import (
	"context"

	Dtos "autobill-service/internal/application/social/dtos"
	Helpers "autobill-service/pkg/helpers"

	"github.com/google/uuid"
)

type SocialUseCase interface {
	GetFriendRequestsList(ctx context.Context, userId uuid.UUID, requestType Dtos.RequestType, pagination Helpers.PaginationParams) (*Dtos.FriendRequestListResult, error)
	SendFriendRequest(ctx context.Context, senderId, receiverId uuid.UUID) (*Dtos.FriendRequestResult, error)
	AcceptFriendRequest(ctx context.Context, senderId, requestId uuid.UUID) error
	RejectFriendRequest(ctx context.Context, senderId, requestId uuid.UUID) error

	GetFriendsList(ctx context.Context, userId uuid.UUID, pagination Helpers.PaginationParams) (*Dtos.FriendsListResult, error)
	RemoveFriend(ctx context.Context, userId, friendId uuid.UUID) error
}
