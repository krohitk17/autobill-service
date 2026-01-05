package Domain

import (
	"github.com/google/uuid"
)

type Friendship struct {
	BaseModel

	UserID   uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_user_friend"`
	FriendID uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_user_friend"`

	User   User `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
	Friend User `gorm:"foreignKey:FriendID;references:Id;constraint:OnDelete:CASCADE"`
}
