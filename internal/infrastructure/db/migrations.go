package DB

import (
	Domain "autobill-service/internal/domain"

	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) {

	db.AutoMigrate(
		&Domain.User{},
		&Domain.Credential{},
		&Domain.RefreshToken{},
		&Domain.Friendship{},
		&Domain.FriendRequest{},
		&Domain.Group{},
		&Domain.GroupMembership{},
		&Domain.Split{},
		&Domain.SplitParticipant{},
		&Domain.ReversalSplit{},
		&Domain.Settlement{},
		&Domain.UserBalance{},
		&Domain.GroupBalance{},
		&Domain.AuditLog{},
	)
}
