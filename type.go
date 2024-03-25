package main

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the structure of a login request.
type LoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the structure of a login response.
type LoginResponse struct {
	Token    string `json:"token"`
	UserName string `json:"userName"`
}

// AccountRequest represents the structure of an account creation request.
type AccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	RoleId    int    `json:"roleId"`
	Country   string `json:"country"`
}

// Account represents the structure of an account.
type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Email             string    `json:"email"`
	Username          string    `json:"username"`
	EncryptedPassword string    `json:"-"`
	Country           string    `json:"country"`
	RoleID            int       `json:"-"`
	CreatedAt         time.Time `json:"createdAt"`
}

// NewAccount creates a new account with the provided details.
func NewAccount(firstname, lastname, email, username, password, country string, roleId int) (*Account, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		FirstName:         firstname,
		LastName:          lastname,
		Email:             email,
		Username:          username,
		EncryptedPassword: string(hash),
		Country:           country,
		RoleID:            roleId,
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// ValidPassword checks if the provided password matches the account's encrypted password.
func (account *Account) ValidPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(account.EncryptedPassword), []byte(password)) == nil
}
