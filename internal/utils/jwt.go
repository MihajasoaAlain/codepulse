package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(username string) (string, error) {

	secret := os.Getenv("JWT_SECRET")

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user": username,
			"exp":  time.Now().Add(time.Hour * 24).Unix(),
		},
	)

	return token.SignedString([]byte(secret))
}
