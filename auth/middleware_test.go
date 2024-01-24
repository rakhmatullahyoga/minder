package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateJWT(t *testing.T) {
	// Create a fake handler to be passed to validateJWT
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Test case: Authorization header is missing, should return 401 Unauthorized
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := ValidateJWT(fakeHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("validateJWT returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	// Test case: JWT is invalid, should return 401 Unauthorized
	req, err = http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer invalid_token")
	rr = httptest.NewRecorder()
	handler = ValidateJWT(fakeHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("validateJWT returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	// Test case: JWT signature is invalid, should return 401 Unauthorized
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		string(ClaimsKeyUserID):   "123",
		string(ClaimsKeyVerified): true,
	}).SignedString([]byte("invalid_key"))
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	handler = ValidateJWT(fakeHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("validateJWT returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	// Test case: JWT is valid, should pass the request to the next handler
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		string(ClaimsKeyUserID):   "123",
		string(ClaimsKeyVerified): true,
	}).SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	handler = ValidateJWT(fakeHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("validateJWT returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
