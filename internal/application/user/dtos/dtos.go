package UserApplicationDtos

import "time"

type UserResult struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateUserInput struct {
	Email string
	Name  string
}

type FindUserByEmailInput struct {
	Email string
}
