package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	if got := Router(); got == nil {
		t.Error("Router() should not return nil")
	}
}

func Test_registerUserHandler(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name           string
		args           args
		init           func()
		responseStatus int
		responseBody   Response
	}{
		{
			name: "unparsed request",
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`broken request`)),
			},
			init:           func() {},
			responseStatus: http.StatusBadRequest,
			responseBody: Response{
				Message: ErrParseInput.Error(),
			},
		},
		{
			name: "invalid request parameter",
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"name":"Yoga"}`)),
			},
			init:           func() {},
			responseStatus: http.StatusBadRequest,
			responseBody: Response{
				Message: ErrInvalidInput.Error(),
			},
		},
		{
			name: "service error",
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"name":"Yoga", "email": "yoga@mail.com", "password":"password"}`)),
			},
			init: func() {
				registerUserService = func(ctx context.Context, in Registration) (err error) {
					err = ErrUserAlreadyRegistered
					return
				}
			},
			responseStatus: http.StatusConflict,
			responseBody: Response{
				Message: ErrUserAlreadyRegistered.Error(),
			},
		},
		{
			name: "internal service error",
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"name":"Yoga", "email": "yoga@mail.com", "password":"password"}`)),
			},
			init: func() {
				registerUserService = func(ctx context.Context, in Registration) (err error) {
					err = ErrUnexpected
					return
				}
			},
			responseStatus: http.StatusInternalServerError,
			responseBody: Response{
				Message: "internal server error",
			},
		},
		{
			name: "success",
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"name":"Yoga", "email": "yoga@mail.com", "password":"password"}`)),
			},
			init: func() {
				registerUserService = func(ctx context.Context, in Registration) (err error) {
					return
				}
			},
			responseStatus: http.StatusCreated,
			responseBody: Response{
				Message: "user successfully registered",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.init()
			registerUserHandler(w, tt.args.r)
			res := w.Result()
			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}

			if res.StatusCode != tt.responseStatus {
				t.Errorf("wrong http status got: %v, want: %v", res.StatusCode, tt.responseStatus)
			}

			resp := Response{}
			json.Unmarshal(data, &resp)
			if resp != tt.responseBody {
				t.Errorf("wrong http response body got: %v, want: %v", resp, tt.responseBody)
			}
		})
	}
}
