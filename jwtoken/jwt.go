package jwtoken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JwtSecret = []byte("your_secret_key")

type UserClaims struct {
	UserID    uint   `json:"user_id"`
	Status    string `json:"status"`
	jwt.RegisteredClaims
}

func GenerateUserJWT(ID uint, IsBlocked bool, Status string) (string, error) {
	fmt.Printf("Signing Secret: %s\n", JwtSecret)

	claims := UserClaims{
		UserID:    ID,
		Status:    Status,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

func GenerateAdminJWT(ID uint, Status string) (string, error) {

	claims := jwt.MapClaims{
		"admin_id": ID,
		"admin_status": Status,
		"iat":     jwt.NewNumericDate(time.Now()),
		"exp":     jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)

}
