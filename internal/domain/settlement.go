package Domain

import (
	"time"

	"github.com/google/uuid"
)

type Settlement struct {
	BaseModel

	SplitID        uuid.UUID `gorm:"type:uuid;index;not null" json:"split_id"`
	PayerID        uuid.UUID `gorm:"type:uuid;index;not null" json:"payer_id"`
	PayeeID        uuid.UUID `gorm:"type:uuid;index;not null" json:"payee_id"`
	Amount         int64     `gorm:"not null" json:"amount"`
	Currency       Currency  `gorm:"type:varchar(10);not null" json:"currency"`
	Date           time.Time `gorm:"not null" json:"date"`
	IdempotencyKey *string   `gorm:"type:varchar(64);uniqueIndex" json:"idempotency_key,omitempty"`

	Split Split `gorm:"foreignKey:SplitID;references:Id;constraint:OnDelete:CASCADE"`
	Payer User  `gorm:"foreignKey:PayerID;references:Id;constraint:OnDelete:CASCADE"`
	Payee User  `gorm:"foreignKey:PayeeID;references:Id;constraint:OnDelete:CASCADE"`
}
