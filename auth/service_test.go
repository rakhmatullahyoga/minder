package auth

import (
	"context"
	"testing"
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
