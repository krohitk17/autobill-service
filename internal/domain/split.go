package Domain

import (
	"github.com/google/uuid"
)

type Split struct {
	BaseModel

	Type          SplitType         `gorm:"type:varchar(20);not null" json:"type"`
	DivisionType  SplitDivisionType `gorm:"type:varchar(20);not null" json:"division_type"`
	TotalAmount   int64             `gorm:"not null" json:"total_amount"`
	Currency      Currency          `gorm:"type:varchar(10);not null" json:"currency"`
	Description   string            `gorm:"type:varchar(500)" json:"description"`
	IsFinalized   bool              `gorm:"default:false" json:"is_finalized"`
	SimplifyDebts *bool             `gorm:"default:null" json:"simplify_debts"`

	GroupID *uuid.UUID `gorm:"type:uuid;index" json:"group_id,omitempty"`
	Group   *Group     `gorm:"foreignKey:GroupID;references:Id;constraint:OnDelete:SET NULL"`

	CreatedByID uuid.UUID `gorm:"type:uuid;index;not null" json:"created_by_id"`
	CreatedBy   User      `gorm:"foreignKey:CreatedByID;references:Id;constraint:OnDelete:CASCADE"`

	Participants []SplitParticipant `gorm:"foreignKey:SplitID;references:Id"`
	Settlements  []Settlement       `gorm:"foreignKey:SplitID;references:Id"`
}
