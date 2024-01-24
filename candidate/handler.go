package candidate

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Response struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	getCandidateFeedService = getCandidateFeed
	swipeCandidateService   = swipeCandidate
	subscribePremiumService = subscribePremium
	getUserInterestService  = getUserInterest
)

func Router() *chi.Mux {
	r := chi.NewMux()

	r.Get("/feed", getFeedHandler)
	r.Post("/swipe", swipeHandler)
	r.Post("/subscribe", subscribeHandler)
	r.Get("/interests", getInterestsHandler)
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
	case ErrInvalidInput:
		httpStatus = http.StatusBadRequest
	case ErrNoCandidateAvailable, ErrAlreadySwiped:
		httpStatus = http.StatusUnprocessableEntity
	case ErrExceedQuota:
		httpStatus = http.StatusPaymentRequired
	case ErrAlreadyVerified:
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

func getFeedHandler(w http.ResponseWriter, r *http.Request) {
	candidate, err := getCandidateFeedService(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Data: candidate}, http.StatusOK)
}

func swipeHandler(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query()
	idStr := val.Get("candidate_id")
	candidateID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, ErrInvalidInput)
		return
	}

	likedStr := val.Get("liked")
	liked, _ := strconv.ParseBool(likedStr)

	candidate, err := swipeCandidateService(r.Context(), candidateID, liked)
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Data: candidate}, http.StatusOK)
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := subscribePremiumService(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Message: "premium subscription success"}, http.StatusOK)
}

func getInterestsHandler(w http.ResponseWriter, r *http.Request) {
	candidates, err := getUserInterestService(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeResponse(w, Response{Data: candidates}, http.StatusOK)
}
