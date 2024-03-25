package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// generateJWT generates a JWT token for the given account.
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

// permissionDenied sends a permission denied response with the given message.
func permissionDenied(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, message)
}

// validateToken validates the JWT token from the request header.
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
