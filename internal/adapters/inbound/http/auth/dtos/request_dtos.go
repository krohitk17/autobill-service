package AuthDtos

type RegisterUserRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"min=8"`
}

type FindUserRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"min=8"`
}

type RefreshTokenRequestDto struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequestDto struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UpdatePasswordRequestDto struct {
	OldPassword string `json:"old_password" validate:"min=8"`
	NewPassword string `json:"new_password" validate:"min=8"`
}

type DeactivateUserRequestDto struct {
	Password string `json:"password" validate:"min=8"`
}

type ReactivateUserRequestDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"min=8"`
}
