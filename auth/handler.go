package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gopkg.in/validator.v2"
)

type Response struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	loginService        = loginUser
	registerUserService = registerUser
)

func Router() *chi.Mux {
	r := chi.NewMux()

	r.Post("/register", registerUserHandler)
	r.Post("/login", loginHandler)
	return r
}

func writeResponse(w http.ResponseWriter, res Response, status int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(res)
}

func mapError(err error) (res Response, httpStatus int) {
	res = Response{
		Message: err.Error(),
	}
	switch err {
	case ErrParseInput, ErrInvalidInput:
		httpStatus = http.StatusBadRequest
	case ErrIncorrectCredentials, ErrUnauthorizedRequest:
		httpStatus = http.StatusUnauthorized
	case ErrUserAlreadyRegistered:
		httpStatus = http.StatusConflict
	default:
		res.Message = "internal server error"
		httpStatus = http.StatusInternalServerError
	}
	return
}

func writeError(w http.ResponseWriter, err error) {
	res, status := mapError(err)
	writeResponse(w, res, status)
}

func registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var params Registration
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, ErrParseInput)
		return
	}

	if err := validator.Validate(params); err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	if err := registerUserService(r.Context(), params); err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Message: "user successfully registered"}, http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var params LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, ErrParseInput)
		return
	}

	if err := validator.Validate(params); err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	res, err := loginService(r.Context(), params)
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Data: res}, http.StatusOK)
}
