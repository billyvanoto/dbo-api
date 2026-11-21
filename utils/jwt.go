package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenDetail struct {
	Token     string `json:"token"`
	Type      string `json:"type"`
	ExpiresIn int    `json:"expiresIn"`
}

const DefaultType = "Bearer"

var DefaultTokenLife = time.Now().Add(24 * time.Hour).Unix()

func GenerateToken(userID uint, email string) (TokenDetail, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     DefaultTokenLife,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return TokenDetail{}, err
	}

	return TokenDetail{
		Token:     tokenStr,
		Type:      DefaultType,
		ExpiresIn: int(DefaultTokenLife),
	}, nil
}
