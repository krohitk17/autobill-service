package Domain

type Group struct {
	BaseModel

	Name          string `gorm:"type:varchar(100);not null" json:"name"`
	SimplifyDebts bool   `gorm:"default:false" json:"simplify_debts"`

	Memberships []GroupMembership `gorm:"foreignKey:GroupID;references:Id"`
	Splits      []Split           `gorm:"foreignKey:GroupID;references:Id"`
	Balances    []GroupBalance    `gorm:"foreignKey:GroupID;references:Id"`
}
