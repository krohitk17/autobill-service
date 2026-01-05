package Domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	BaseModel

	UserID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Action     string     `gorm:"type:varchar(255);not null" json:"action"`
	EntityType string     `gorm:"type:varchar(50);index" json:"entity_type"`
	EntityID   *uuid.UUID `gorm:"type:uuid;index" json:"entity_id,omitempty"`
	GroupID    *uuid.UUID `gorm:"type:uuid;index" json:"group_id,omitempty"`
	SplitID    *uuid.UUID `gorm:"type:uuid;index" json:"split_id,omitempty"`
	Details    string     `gorm:"type:text" json:"details,omitempty"`
	Timestamp  time.Time  `gorm:"not null" json:"timestamp"`

	User User `gorm:"foreignKey:UserID;references:Id;constraint:OnDelete:CASCADE"`
}
