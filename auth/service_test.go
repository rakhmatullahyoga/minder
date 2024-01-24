package auth

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-chi/jwtauth/v5"
)

func Test_registerUser(t *testing.T) {
	type args struct {
		ctx context.Context
		in  Registration
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		errMessage string
		init       func()
	}{
		{
			name: "error when checking user email existed",
			args: args{
				ctx: context.Background(),
				in:  Registration{Name: "Yoga", Password: "password", Email: "yoga@mail.com"},
			},
			wantErr:    true,
			errMessage: ErrUnexpected.Error(),
			init: func() {
				userEmailExistedRepo = func(ctx context.Context, email string) (existed bool, err error) {
					err = ErrUnexpected
					return
				}
			},
		},
		{
			name: "user email already used",
			args: args{
				ctx: context.Background(),
				in:  Registration{Name: "Yoga", Password: "password", Email: "yoga@mail.com"},
			},
			wantErr:    true,
			errMessage: ErrUserAlreadyRegistered.Error(),
			init: func() {
				userEmailExistedRepo = func(ctx context.Context, email string) (existed bool, err error) {
					existed = true
					return
				}
			},
		},
		{
			name: "error saving user to database",
			args: args{
				ctx: context.Background(),
				in:  Registration{Name: "Yoga", Password: "password", Email: "yoga@mail.com"},
			},
			wantErr:    true,
			errMessage: ErrUnexpected.Error(),
			init: func() {
				userEmailExistedRepo = func(ctx context.Context, email string) (existed bool, err error) {
					return
				}
				createUserRepo = func(ctx context.Context, in Registration) (err error) {
					err = ErrUnexpected
					return
				}
			},
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				in:  Registration{Name: "Yoga", Password: "password", Email: "yoga@mail.com"},
			},
			init: func() {
				userEmailExistedRepo = func(ctx context.Context, email string) (existed bool, err error) {
					return
				}
				createUserRepo = func(ctx context.Context, in Registration) (err error) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init()
			err := registerUser(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("wrong error message got: %s, want: %s", err.Error(), tt.errMessage)
			}
		})
	}
}

func Test_loginUser(t *testing.T) {
	secretKey := "super-secret"
	req := LoginRequest{
		Email:    "yoga@mail.com",
		Password: "password",
	}
	type args struct {
		ctx context.Context
		req LoginRequest
	}
	tests := []struct {
		name       string
		args       args
		wantRes    LoginResponse
		wantErr    bool
		errMessage string
		init       func()
	}{
		{
			name: "unexpected error when getting user",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			wantErr:    true,
			errMessage: ErrUnexpected.Error(),
			init: func() {
				getUserByEmailRepo = func(ctx context.Context, email string) (user User, err error) {
					err = ErrUnexpected
					return
				}
			},
		},
		{
			name: "incorrect email",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			wantErr:    true,
			errMessage: ErrIncorrectCredentials.Error(),
			init: func() {
				getUserByEmailRepo = func(ctx context.Context, email string) (user User, err error) {
					err = ErrUserNotFound
					return
				}
			},
		},
		{
			name: "incorrect password",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			wantErr:    true,
			errMessage: ErrIncorrectCredentials.Error(),
			init: func() {
				getUserByEmailRepo = func(ctx context.Context, email string) (user User, err error) {
					user = User{
						ID:       1,
						Name:     "Yoga",
						Email:    req.Email,
						Password: hashPassword("pasword"),
						Verified: true,
					}
					return
				}
			},
		},
		{
			name: "error generating token",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			wantErr: true,
			init: func() {
				TokenAuth = jwtauth.New("abc", []byte("JWT_SECRET_KEY"), nil)
				getUserByEmailRepo = func(ctx context.Context, email string) (user User, err error) {
					user = User{
						ID:       1,
						Name:     "Yoga",
						Email:    req.Email,
						Password: hashPassword(req.Password),
						Verified: true,
					}
					return
				}
			},
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				req: req,
			},
			wantErr: false,
			wantRes: LoginResponse{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InlvZ2FAbWFpbC5jb20iLCJpZCI6MSwibmFtZSI6IllvZ2EiLCJ2ZXJpZmllZCI6dHJ1ZX0.65ss2uvz8-8QyQtsb8WR0gXo9aglvPwPj3EgOL212M8",
			},
			init: func() {
				TokenAuth = jwtauth.New("HS256", []byte(secretKey), nil)
				getUserByEmailRepo = func(ctx context.Context, email string) (user User, err error) {
					user = User{
						ID:       1,
						Name:     "Yoga",
						Email:    req.Email,
						Password: hashPassword(req.Password),
						Verified: true,
					}
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init()
			gotRes, err := loginUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("loginUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMessage != "" && err.Error() != tt.errMessage {
				t.Errorf("wrong error message got: %s, want: %s", err.Error(), tt.errMessage)
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("loginUser() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
