package UserDtos

type UpdateUserRequestDto struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

type FindUserByEmailRequestDto struct {
	Email string `json:"email" validate:"required,email"`
}
