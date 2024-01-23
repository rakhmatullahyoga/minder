package auth

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

func registerUser(ctx context.Context, in Registration) (err error) {
	existed, err := isUserEmailExisted(ctx, in.Email)
	if err != nil {
		return
	}
	if existed {
		err = ErrUserAlreadyRegistered
		return
	}

	hashPassword(&in)
	err = createUser(ctx, in)
	return
}

func hashPassword(in *Registration) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	in.Password = string(hash)
}
