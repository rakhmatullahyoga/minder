package auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

var (
	userEmailExistedRepo = isUserEmailExisted
	createUserRepo       = createUser
)

func registerUser(ctx context.Context, in Registration) (err error) {
	existed, err := userEmailExistedRepo(ctx, in.Email)
	if err != nil {
		return
	}
	if existed {
		err = ErrUserAlreadyRegistered
		return
	}

	hashPassword(&in)
	err = createUserRepo(ctx, in)
	return
}

func hashPassword(in *Registration) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	in.Password = string(hash)
}
