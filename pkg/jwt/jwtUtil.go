package JWTUtil

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil struct {
	Secret                 string
	Expiration             time.Duration
	RefreshTokenExpiration time.Duration
}

func CreateJwtUtil(secret string, expiration, refreshExpiration time.Duration) JWTUtil {
	return JWTUtil{Secret: secret, Expiration: expiration, RefreshTokenExpiration: refreshExpiration}
}

func (j JWTUtil) Generate(userID string) (string, error) {
	claimsDto := JWTClaims{
		Id:  userID,
		Exp: time.Now().Add(j.Expiration).Unix(),
	}
	claims := jwt.MapClaims(claimsDto.ToMap())
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(j.Secret))
}

func (j JWTUtil) Parse(token string) (string, bool) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.Secret), nil
	})

	if err != nil || !t.Valid {
		return "", false
	}

	claims := t.Claims.(jwt.MapClaims)
	claimsDto := &JWTClaims{}
	claimsDto.FromMap(claims)
	return claimsDto.Id, true
}

func (j JWTUtil) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (j JWTUtil) GetRefreshTokenExpiry() time.Time {
	return time.Now().Add(j.RefreshTokenExpiration)
}
