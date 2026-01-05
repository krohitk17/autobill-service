package Domain

import (
	"github.com/google/uuid"
)

type Credential struct {
	BaseModel

	UserID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null;size:255;check:char_length(password_hash) >= 8"`
}
