package Domain

import (
	"github.com/google/uuid"
)

type ReversalSplit struct {
	BaseModel

	OriginalSplitID uuid.UUID `gorm:"type:uuid;index;not null" json:"original_split_id"`
	ReversalSplitID uuid.UUID `gorm:"type:uuid;index;not null" json:"reversal_split_id"`
	Reason          string    `gorm:"type:varchar(500)" json:"reason"`

	OriginalSplit Split `gorm:"foreignKey:OriginalSplitID;references:Id;constraint:OnDelete:CASCADE"`
	ReversalSplit Split `gorm:"foreignKey:ReversalSplitID;references:Id;constraint:OnDelete:CASCADE"`
}
