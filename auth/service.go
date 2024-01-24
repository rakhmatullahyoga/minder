package auth

import (
	"context"
	"os"

	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	tokenAuth = jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET_KEY")), nil)

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

	token, err := generateToken(user)
	if err != nil {
		return
	}

	res.Token = token
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

func generateToken(user User) (token string, err error) {
	claims := map[string]interface{}{
		"id":       user.ID,
		"email":    user.Email,
		"name":     user.Name,
		"verified": user.Verified,
	}
	_, token, err = tokenAuth.Encode(claims)
	return
}
