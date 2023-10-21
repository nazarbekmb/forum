package models

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword string
}

// Define a new UserModel type which wraps a database connection pool.
type UserModel struct {
	DB *sql.DB
}

// We'll use the Insert method to add a new record to the "users" table.
func (u *UserModel) Insert(name, email, hashPassword string) error {
	// db, err := sql.Open("sqlite3", "./forum.db")
	// defer db.Close()
	stmt := `INSERT INTO users (username, email, hash_password)
	VALUES ($1, $2, $3)`
	_, err := u.DB.Exec(stmt, name, email, hashPassword)
	if err != nil {
		if err != nil {
			if sqliteErr, ok := err.(sqlite3.Error); ok {
				if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(sqliteErr.Error(), "users_uc_email") {
					return ErrDuplicateEmail
				}
				if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(sqliteErr.Error(), "users_uc_username") {
					return ErrDuplicateUsername
				}
			}
			return err
		}
	}
	return nil
}

func (u *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := "SELECT user_id, hash_password FROM users WHERE email = ?"
	err := u.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, nil
}
