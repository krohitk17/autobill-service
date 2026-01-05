package Domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	BaseModel

	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`

	User User `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && time.Now().Before(rt.ExpiresAt)
}
