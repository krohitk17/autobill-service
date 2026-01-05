package AuthDtos

type UserLoginResponseDto struct {
	Id           string `json:"id"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
