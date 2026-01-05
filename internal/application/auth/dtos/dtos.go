package AuthApplicationDtos

type RegisterUserInput struct {
	Email    string
	Name     string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshTokenInput struct {
	RefreshToken string
}

type UpdatePasswordInput struct {
	OldPassword string
	NewPassword string
}

type DeactivateUserInput struct {
	Password string
}

type ReactivateUserInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	ID           string
	Token        string
	RefreshToken string
}
