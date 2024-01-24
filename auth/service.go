package auth

import (
	"context"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	createUserRepo       = createUser
	getUserByEmailRepo   = getUserByEmail
	userEmailExistedRepo = isUserEmailExisted
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

	in.Password = hashPassword(in.Password)
	err = createUserRepo(ctx, in)
	return
}

func loginUser(ctx context.Context, req LoginRequest) (res LoginResponse, err error) {
	user, err := getUserByEmailRepo(ctx, req.Email)
	if err != nil {
		if err == ErrUserNotFound {
			err = ErrIncorrectCredentials
		}
		return
	}

	if !checkPassword(user.Password, req.Password) {
		err = ErrIncorrectCredentials
		return
	}

	res.Token = generateToken(user)
	return
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func checkPassword(hashPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil
}

func generateToken(user User) (token string) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"verified": user.Verified,
	})
	token, _ = tkn.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	return
}
