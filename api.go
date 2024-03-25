package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"time"
)

// APIServer represents the API server.
type APIServer struct {
	listenAddr  string        // Address to listen on
	dbStore     *PostgresDB   // Database store
	redisClient *redis.Client // Redis client
}

// apiFunc is a function type for handling API requests.
type apiFunc func(w http.ResponseWriter, r *http.Request) error

// ApiError represents an API error response.
type ApiError struct {
	Error string `json:"error"`
}

// makeHTTPHandleFunc creates an HTTP handler function from an apiFunc.
func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := f(writer, request); err != nil {
			writeJSON(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

// writeJSON writes JSON data to the response writer.
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// Run starts the API server.
func (s *APIServer) Run() error {
	router := mux.NewRouter()

	// Swagger endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(httpSwagger.URL("/docs/swagger.json")))
	router.HandleFunc("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})
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

// newAPIServer creates a new APIServer instance.
func newAPIServer(listenAddr string, store *PostgresDB, redisClient *redis.Client) *APIServer {
	return &APIServer{
		listenAddr:  listenAddr,
		dbStore:     store,
		redisClient: redisClient,
	}
}

// handleLogin handles the login request.
// Login endpoint
// @Summary Log in with username and password
// @Tags auth
// @Accept json
// @Produce json
// username body string true "UserName"
// @Param request body LoginRequest true "Login details"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ApiError
// @Router /login [post]
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

// handleLogout handles the logout request.
// Logout endpoint
// @Summary Log out
// @Tags auth
// @Produce plain
// @Param id path int true "Account ID"
// @Param token header string true "Auth token"
// @Success 200 {string} string
// @Router /{id}/logout [get]
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

// handleGetAllAccount handles the request to get all accounts.
// @Summary Get all accounts.
// @Description Retrieves a list of all accounts.
// @Produce json
// @Success 200 {array} Account
// @Router /account [get]
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

// handleGetAccountByID handles the request to get an account by ID.
// @Summary Get account by ID
// @Tags accounts
// @Produce json
// @Param id path int true "Account ID"
// @Param token header string true "Auth token"
// @Success 200 {object} Account
// @Failure 404 {object} ApiError
// @Router /account/{id} [get]
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}

		account, err := s.dbStore.GetAccountByID(id)
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

// handleCreateAccount handles the request to create an account.
// @Summary Create a new account.
// @Description Creates a new account based on the provided request data.
// @Accept json
// @Produce json
// @Param request body AccountRequest true "Account details to create"
// @Success 200 {object} Account
// @Router /account [post]
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

// @Summary Delete an account by ID
// @Description Deletes an account by its ID
// @Tags accounts
// @Accept json
// @Produce json
// @Param token header string true "Auth token"
// @Param id path int true "Account ID"
// @Success 200 {object} map[string]int "deleted":int "Success"
// @Router /account/{id} [delete]
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
