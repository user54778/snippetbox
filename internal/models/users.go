package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	PasswordUpdate(id int, currentPassword, newPassword string) error
}

// A new user type to directly represent the database.
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

// Wrap the database connection pool.
type UserModel struct {
	DB *sql.DB
}

// Now, we will define methods on this type
// for interacting with the Users database.
func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES(?, ?, ?, UTC_TIMESTAMP())`
	// Use the Exec() method to insert the user details and hashed password
	// into the users table.
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// If this returns an error, we use the errors.As() function to check
		// whether the error has the type *mysql.MySQLError. If it does, the
		// error will be assigned to the mySQLError variable. We can then check
		// whether or not the error relates to our users_uc_email key by
		// checking if the error code equals 1062 and the contents of the error
		// message string. If it does, we return an ErrDuplicateEmail error.
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

// Authenticate a users existence with provided email and
// password in the database.
// Return the user's ID if they do exist.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// 1) Retrieve the hashed password associated with the email address.
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"

	// Scan() copies the columns returned by QueryRow().
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		// Doesn't exists or user has been deactivated
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, nil
		}
	}

	// 2) Compare the hashed password with the plain-text password the user
	// provided when logging in.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		// Invalid password
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Password is correct. Return the user ID.
	return id, nil
}

// Check if a user exists with a given ID.
func (m *UserModel) Exists(id int) (bool, error) {
	// NOTE: updated to return true if user w/ specific ID exists
	// in our users table, false otherwise.
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"

	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

// Get a user id from the `users` database.
func (m *UserModel) Get(id int) (*User, error) {
	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`

	tuple := m.DB.QueryRow(stmt, id)

	// zeroed User pointer
	user := &User{}

	err := tuple.Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	if err != nil {
		// No tuples returned
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return user, nil
}

// Updates a user's password in the `users` database.
func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	// retrieve the user's details via id from the database.
	var hashedPassword []byte
	stmt := `SELECT hashed_password FROM users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&hashedPassword)
	if err != nil {
		// Doesn't exists or user has been deactivated
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		} else {
			return nil
		}
	}
	// check the currentPassword matches the hashed password.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		} else {
			return err
		}
	}
	// hash the newPassword value
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	// update the hashed_password column in the users table
	stmt = `UPDATE users SET hashed_password = ? WHERE id = ?`

	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)

	return err
}
