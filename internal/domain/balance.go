package Domain

import (
	"github.com/google/uuid"
)

type UserBalance struct {
	BaseModel

	UserID      uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	OtherUserID uuid.UUID `gorm:"type:uuid;index;not null" json:"other_user_id"`
	NetAmount   int64     `gorm:"not null;default:0" json:"net_amount"`
	Currency    Currency  `gorm:"type:varchar(10);not null" json:"currency"`

	User      User `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
	OtherUser User `gorm:"foreignKey:OtherUserID;references:Id;constraint:OnDelete:CASCADE"`
}

type GroupBalance struct {
	BaseModel

	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	GroupID   uuid.UUID `gorm:"type:uuid;index;not null" json:"group_id"`
	NetAmount int64     `gorm:"not null;default:0" json:"net_amount"`
	Currency  Currency  `gorm:"type:varchar(10);not null" json:"currency"`

	User  User  `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
	Group Group `gorm:"foreignKey:GroupID;references:Id;constraint:OnDelete:CASCADE"`
}

type SimplifiedDebt struct {
	FromUserID   uuid.UUID `json:"from_user_id"`
	FromUserName string    `json:"from_user_name"`
	ToUserID     uuid.UUID `json:"to_user_id"`
	ToUserName   string    `json:"to_user_name"`
	Amount       int64     `json:"amount"`
	Currency     Currency  `json:"currency"`
}
