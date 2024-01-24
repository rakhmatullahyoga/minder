package auth

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ClaimsKey string

const (
	ClaimsKeyUserID   ClaimsKey = "user_id"
	ClaimsKeyVerified ClaimsKey = "verified"
)

func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) < 2 {
			writeError(w, ErrUnauthorizedRequest)
			return
		}
		reqToken = splitToken[1]
		token, err := jwt.Parse(reqToken, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil {
			writeError(w, ErrUnauthorizedRequest)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := r.Context()
			ctx = context.WithValue(ctx, ClaimsKeyUserID, claims["id"])
			ctx = context.WithValue(ctx, ClaimsKeyVerified, claims["verified"])
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			writeError(w, ErrUnauthorizedRequest)
			return
		}
	})
}
