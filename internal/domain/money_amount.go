package Domain

type MoneyAmount struct {
	Value    int64    `gorm:"not null" json:"value"`
	Currency Currency `gorm:"type:varchar(10);not null" json:"currency"`
}
