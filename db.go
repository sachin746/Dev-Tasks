package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	GetAllAccount() ([]*Account, error)
	GetAccountById(int) (*Account, error)
	DeleteAccount(int) error
	UpdateAccount(*Account) error
}

type PostgresDB struct {
	db *sql.DB
}

// host=host.docker.internal
func NewPostgresDB() (*PostgresDB, error) {
	connStr := " user=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &PostgresDB{
		db: db,
	}, nil
}

func (s *PostgresDB) initDB() error {
	err := s.CreateRoleTable()
	if err != nil {
		return err
	}
	err = s.CreateAccountTable()
	if err != nil {
		return err
	}
	return err
}

func (s *PostgresDB) CreateAccountTable() error {
	Query := `create table if not exists account (
    			id serial Primary Key,
    			firstName varchar(255),
    			lastName  varchar(255),
    			email varchar(255),
    			username varchar(255),
    			hash varchar(255),
    			country varchar(255),
    			roleId int references role(id),
    			createdAt timestamp
    )`
	_, err := s.db.Exec(Query)
	return err
}
func (s *PostgresDB) CreateRoleTable() error {
	Query := `CREATE TABLE if not exists role (
    	id SERIAL PRIMARY KEY,
    	name VARCHAR(50) UNIQUE
	);`
	_, err := s.db.Exec(Query)
	var rowCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM role").Scan(&rowCount)
	if err != nil {
		panic(err)
	}

	// Check if the table is empty
	if rowCount == 0 {
		QueryForRoles := "insert into role (name) values ('admin'), ('user')"
		_, err = s.db.Query(QueryForRoles)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *PostgresDB) CreateAccount(account *Account) error {
	Query := `insert into account (firstName, lastName, email, username, hash, country, roleId, createdAt)
				values ($1,$2, $3, $4, $5, $6, $7, $8)`
	_, err := s.db.Query(Query,
		account.FirstName,
		account.LastName,
		account.Email,
		account.Username,
		account.EncryptedPassword,
		account.Country,
		account.RoleID,
		account.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
func (s *PostgresDB) GetAllAccounts() ([]*Account, error) {
	Query := `select * from account`
	rows, err := s.db.Query(Query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var accounts []*Account
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			panic(err)
		}
		accounts = append(accounts, account)
	}
	return accounts, err
}

func (s *PostgresDB) GetAccountById(id int) (*Account, error) {
	rows, err := s.db.Query("select * from account where id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}
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

func (s *PostgresDB) DeleteAccount(id int) error {
	_, err := s.db.Query("delete from account where id = $1", id)
	fmt.Println(err)
	return err
}
