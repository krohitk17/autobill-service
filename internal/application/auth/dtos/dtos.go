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

type AuthResult struct {
	ID           string
	Token        string
	RefreshToken string
}
