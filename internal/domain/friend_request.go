package Domain

import (
	"github.com/google/uuid"
)

type FriendRequest struct {
	BaseModel

	SenderId   uuid.UUID    `gorm:"type:uuid;index;not null"`
	ReceiverId uuid.UUID    `gorm:"type:uuid;index;not null"`
	Status     FriendStatus `gorm:"type:varchar(20);not null"`

	Sender   User `gorm:"foreignKey:SenderId;references:Id;constraint:OnDelete:CASCADE"`
	Receiver User `gorm:"foreignKey:ReceiverId;references:Id;constraint:OnDelete:CASCADE"`
}

type FriendStatus string

const (
	FriendPending  FriendStatus = "PENDING"
	FriendAccepted FriendStatus = "ACCEPTED"
	FriendRejected FriendStatus = "REJECTED"
)
