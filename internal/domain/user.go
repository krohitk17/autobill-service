package Domain

type User struct {
	BaseModel

	Email  string        `gorm:"uniqueIndex;not null" json:"email"`
	Name   string        `json:"name"`
	Status AccountStatus `gorm:"type:varchar(20);not null" json:"status"`

	Credential Credential `gorm:"foreignKey:UserID;references:Id"`

	SentFriendRequests     []FriendRequest `gorm:"foreignKey:SenderID;references:Id"`
	ReceivedFriendRequests []FriendRequest `gorm:"foreignKey:ReceiverID;references:Id"`
	Friendships            []Friendship    `gorm:"foreignKey:UserID;references:Id"`

	GroupMemberships  []GroupMembership  `gorm:"foreignKey:UserID;references:Id"`
	SplitParticipants []SplitParticipant `gorm:"foreignKey:UserID;references:Id"`
	UserBalances      []UserBalance      `gorm:"foreignKey:UserID;references:Id"`
	GroupBalances     []GroupBalance     `gorm:"foreignKey:UserID;references:Id"`
	AuditLogs         []AuditLog         `gorm:"foreignKey:UserID;references:Id"`
}

type AccountStatus string

const (
	AccountActive      AccountStatus = "ACTIVE"
	AccountDeactivated AccountStatus = "DEACTIVATED"
)
