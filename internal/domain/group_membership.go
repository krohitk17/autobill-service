package Domain

import (
	"github.com/google/uuid"
)

type GroupMembership struct {
	BaseModel

	UserID  uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_group_user" json:"user_id"`
	GroupID uuid.UUID `gorm:"type:uuid;index;not null;uniqueIndex:idx_group_user" json:"group_id"`
	Role    GroupRole `gorm:"type:varchar(20);not null" json:"role"`

	User  User  `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
	Group Group `gorm:"foreignKey:GroupID;references:Id;constraint:OnDelete:CASCADE"`
}
