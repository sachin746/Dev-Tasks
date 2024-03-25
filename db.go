package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

// Storage defines the methods for interacting with the database.
type Storage interface {
	CreateAccount(*Account) error
	GetAllAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	DeleteAccount(int) error
	UpdateAccount(*Account) error
}

// PostgresDB represents a connection to a PostgreSQL database.
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgresDB instance.
func NewPostgresDB() (*PostgresDB, error) {
	connStr := "user=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresDB{db: db}, nil
}

// InitDB initializes the database schema.
func (s *PostgresDB) InitDB() error {
	if err := s.CreateRoleTable(); err != nil {
		return err
	}
	if err := s.CreateAccountTable(); err != nil {
		return err
	}
	return nil
}

// CreateAccountTable creates the account table if it does not exist.
func (s *PostgresDB) CreateAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
		id SERIAL PRIMARY KEY,
		firstName VARCHAR(255),
		lastName VARCHAR(255),
		email VARCHAR(255),
		username VARCHAR(255),
		hash VARCHAR(255),
		country VARCHAR(255),
		roleID INT REFERENCES role(id),
		createdAt TIMESTAMP
	)`
	_, err := s.db.Exec(query)
	return err
}

// CreateRoleTable creates the role table if it does not exist.
func (s *PostgresDB) CreateRoleTable() error {
	query := `CREATE TABLE IF NOT EXISTS role (
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) UNIQUE
	)`
	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	// Check if the table is empty
	var rowCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM role").Scan(&rowCount)
	if err != nil {
		return err
	}
	if rowCount == 0 {
		// Insert default roles
		query := "INSERT INTO role (name) VALUES ('admin'), ('user')"
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateAccount inserts a new account into the database.
func (s *PostgresDB) CreateAccount(account *Account) error {
	query := `INSERT INTO account (firstName, lastName, email, username, hash, country, roleID, createdAt)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.Exec(query, account.FirstName, account.LastName, account.Email,
		account.Username, account.EncryptedPassword, account.Country, account.RoleID, account.CreatedAt)
	return err
}

// GetAllAccounts retrieves all accounts from the database.
func (s *PostgresDB) GetAllAccounts() ([]*Account, error) {
	query := `SELECT * FROM account`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// GetAccountByID retrieves an account by its ID from the database.
func (s *PostgresDB) GetAccountByID(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

// GetAccountByUsername retrieves an account by its username from the database.
func (s *PostgresDB) GetAccountByUsername(username string) (*Account, error) {
	rows, err := s.db.Query("select * from account where username = $1", username)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %s not found", username)
}

// DeleteAccount deletes an account from the database by its ID.
func (s *PostgresDB) DeleteAccount(id int) error {
	query := `DELETE FROM account WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
