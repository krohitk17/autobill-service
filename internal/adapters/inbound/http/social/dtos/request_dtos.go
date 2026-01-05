package SocialDtos

import "github.com/google/uuid"

type SendFriendRequestRequestDto struct {
	ReceiverId uuid.UUID `json:"receiver_id" validate:"required"`
}
