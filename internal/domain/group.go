package Domain

import "github.com/google/uuid"

type Group struct {
	BaseModel

	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID       uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	SimplifyDebts bool      `gorm:"default:false" json:"simplify_debts"`

	Memberships []GroupMembership `gorm:"foreignKey:GroupID;references:Id"`
	Splits      []Split           `gorm:"foreignKey:GroupID;references:Id"`
	Balances    []GroupBalance    `gorm:"foreignKey:GroupID;references:Id"`
}
