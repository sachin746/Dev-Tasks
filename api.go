package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

type APIServer struct {
	listenAddr  string
	dbStore     *PostgresDB
	redisClient *redis.Client
}
type apiFunc func(w http.ResponseWriter, r *http.Request) error
type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := f(writer, request); err != nil {
			writeJSON(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
	router.HandleFunc("/{id}/logout", isAuthenticated(makeHTTPHandleFunc(s.handleLogout), s))
	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", isAuthenticated(makeHTTPHandleFunc(s.handleGetAccountByID), s))
	log.Printf("Server running on port %v", s.listenAddr)
	err := http.ListenAndServe(s.listenAddr, router)
	if err != nil {
		return fmt.Errorf("starting server run into problems")
	}
	return nil
}
func newAPIServer(listenAddr string, store *PostgresDB, redisClient *redis.Client) *APIServer {
	return &APIServer{
		listenAddr:  listenAddr,
		dbStore:     store,
		redisClient: redisClient,
	}
}
func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	account, err := s.dbStore.GetAccountByUsername(req.UserName)
	if err != nil {
		return err
	}
	if !account.ValidPassword(req.Password) {
		permissionDenied(w, "permission denied")
		return nil
	}
	token, err := generateJWT(account)
	if err != nil {
		return err
	}
	resp := &LoginResponse{
		Token:    token,
		UserName: account.Username,
	}
	return writeJSON(w, http.StatusOK, resp)
}
func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) error {
	// Get token from cookie
	token := r.Header.Get("token")

	// Blacklist token in Redis
	s.redisClient.Set(context.Background(), token, "", 1*time.Minute)

	// Clear token cookie
	r.Header.Del("token")

	w.WriteHeader(http.StatusOK)
	return writeJSON(w, http.StatusOK, "Logout successful")
}
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAllAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}
	return fmt.Errorf("%v Method not Allow", r.Method)
}

func (s *APIServer) handleGetAllAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.dbStore.GetAllAccounts()
	if err != nil {
		return err
	}
	err = writeJSON(w, http.StatusOK, accounts)
	if err != nil {
		return err
	}
	return nil
}
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}

		account, err := s.dbStore.GetAccountById(id)
		if err != nil {
			return err
		}
		return writeJSON(w, http.StatusOK, account)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	var accountReq *AccountRequest
	err := json.NewDecoder(r.Body).Decode(&accountReq)
	if err != nil {
		return err
	}
	if accountReq.RoleId == 0 {
		accountReq.RoleId = 2
	}
	account, err := NewAccount(
		accountReq.FirstName,
		accountReq.LastName,
		accountReq.Email,
		accountReq.Username,
		accountReq.Password,
		accountReq.Country,
		accountReq.RoleId)
	if err != nil {
		return err
	}
	err = s.dbStore.CreateAccount(account)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	if err = s.dbStore.DeleteAccount(id); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, map[string]int{"deleted": id})
}
