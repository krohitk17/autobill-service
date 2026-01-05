package Domain

import (
	"github.com/google/uuid"
)

type SplitParticipant struct {
	BaseModel

	SplitID     uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_split_user" json:"split_id"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_split_user" json:"user_id"`
	ShareAmount int64     `gorm:"not null" json:"share_amount"`
	Currency    Currency  `gorm:"type:varchar(10);not null" json:"currency"`
	IsSettled   bool      `gorm:"not null;default:false" json:"is_settled"`

	Split Split `gorm:"foreignKey:SplitID;references:Id;constraint:OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
}
