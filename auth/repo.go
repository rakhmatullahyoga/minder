package auth

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	db *sqlx.DB
)

const (
	// queries
	checkUserEmailExists = "SELECT EXISTS (SELECT email FROM users WHERE email = ?)"
	insertUserQuery      = "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"

	// others
	duplicateDbConstraint = "Duplicate entry"
)

func SetDB(newdb *sqlx.DB) (err error) {
	if newdb == nil {
		err = ErrEmptyDB
		return
	}

	db = newdb
	return
}

func createUser(ctx context.Context, in Registration) (err error) {
	_, err = db.ExecContext(ctx, insertUserQuery, in.Name, in.Email, in.Password)
	if err != nil {
		if strings.Contains(err.Error(), duplicateDbConstraint) {
			err = ErrUserAlreadyRegistered
		}
	}
	return
}

func isUserEmailExisted(ctx context.Context, email string) (existed bool, err error) {
	err = db.QueryRowContext(ctx, checkUserEmailExists, email).Scan(&existed)
	return
}
