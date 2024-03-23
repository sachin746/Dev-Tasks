package main

import (
	bcrypt2 "golang.org/x/crypto/bcrypt"
	"time"
)

type LoginRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Token    string `json:"token"`
	UserName string `json:"userName"`
}
type AccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	RoleId    int    `json:"roleId"`
	Country   string `json:"country"`
}

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

func NewAccount(firstname, lastname, email, username, password, country string, roleId int) (*Account, error) {
	hash, err := bcrypt2.GenerateFromPassword([]byte(password), bcrypt2.DefaultCost)
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
func (account *Account) ValidPassword(password string) bool {
	return bcrypt2.CompareHashAndPassword([]byte(account.EncryptedPassword), []byte(password)) == nil
}
