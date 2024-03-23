package main

import (
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Email,
		&account.Username,
		&account.EncryptedPassword,
		&account.Country,
		&account.RoleID,
		&account.CreatedAt)
	return account, err
}
func isAuthenticated(handlerFunc http.HandlerFunc, s *APIServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")
		if isBlacklistedToken(w, r, s.redisClient) {
			return
		}
		tokenFromHeader := r.Header.Get("token")
		token, err := validateToken(tokenFromHeader)
		if err != nil {
			permissionDenied(w, "token not verified")
			return
		}
		if !token.Valid {
			permissionDenied(w, "token not valid")
			return
		}
		userID, err := getID(r)
		if err != nil {
			permissionDenied(w, "error fetching id")
			return
		}
		account, err := s.dbStore.GetAccountById(userID)
		if err != nil {
			permissionDenied(w, "error fetching account")
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Username != claims["username"] || account.RoleID != int(claims["role"].(float64)) {
			permissionDenied(w, "unauthorized")
			return
		}

		if err != nil {
			writeJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
}
