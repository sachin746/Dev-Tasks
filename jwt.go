package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
)

func generateJWT(account *Account) (string, error) {
	claims := jwt.MapClaims{
		"exp":      time.Now().Add(1 * time.Minute).Unix(),
		"username": account.Username,
		"role":     account.RoleID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := os.Getenv("bank_secret")
	return token.SignedString([]byte(secretKey))
}

func permissionDenied(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, message)
}

func validateToken(tokenFromHeader string) (*jwt.Token, error) {
	secretKey := []byte(os.Getenv("bank_secret"))
	checkedToken, err := jwt.Parse(tokenFromHeader, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing token.")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("your Token has been expired")
	}
	return checkedToken, nil
}
