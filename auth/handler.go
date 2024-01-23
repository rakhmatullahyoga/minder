package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gopkg.in/validator.v2"
)

type Response struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func Router() *chi.Mux {
	r := chi.NewMux()

	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Post("/register", registerUserHandler)
	r.Post("/login", loginHandler)
	return r
}

func writeResponse(w http.ResponseWriter, res Response, status int) {
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

func mapError(err error) (res Response, httpStatus int) {
	res = Response{
		Message: err.Error(),
	}
	switch err {
	// case ErrParseInput, ErrTitleMissing, ErrSkuMissing, ErrPriceMissing, ErrRatingMissing, ErrRatingInvalid:
	// 	httpStatus = http.StatusBadRequest
	// case ErrNotFound:
	// 	httpStatus = http.StatusNotFound
	// case ErrNoRows, ErrSkuDuplicate:
	// 	httpStatus = http.StatusUnprocessableEntity
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
	var regParams Registration
	if err := json.NewDecoder(r.Body).Decode(&regParams); err != nil {
		writeError(w, ErrParseInput)
		return
	}

	if err := validator.Validate(regParams); err != nil {
		writeError(w, err)
		return
	}

	if err := registerUser(r.Context(), regParams); err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Message: "user successfully registered"}, http.StatusCreated)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {}
