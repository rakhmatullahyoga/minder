package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestSetDB(t *testing.T) {
	mockDb, _, _ := sqlmock.New()
	defer mockDb.Close()
	testDb := sqlx.NewDb(mockDb, "sqlmock")
	type args struct {
		newdb *sqlx.DB
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "empty db",
			args:    args{},
			wantErr: ErrEmptyDB,
		},
		{
			name: "with mock db",
			args: args{
				newdb: testDb,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetDB(tt.args.newdb); err != tt.wantErr {
				t.Errorf("SetDB() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createUser(t *testing.T) {
	input := Registration{Name: "Yoga", Password: "password", Email: "yoga@mail.com"}
	mockDb, mockQuery, _ := sqlmock.New()
	defer mockDb.Close()
	testDb := sqlx.NewDb(mockDb, "sqlmock")
	SetDB(testDb)
	type args struct {
		ctx context.Context
		in  Registration
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		errMessage string
		init       func(sqlmock.Sqlmock)
	}{
		{
			name: "user email already registered",
			args: args{
				ctx: context.Background(),
				in:  input,
			},
			wantErr:    true,
			errMessage: ErrUserAlreadyRegistered.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectExec("INSERT INTO users").WithArgs(input.Name, input.Email, input.Password).WillReturnError(errors.New(duplicateDbConstraint))
			},
		},
		{
			name: "unexpected error from database",
			args: args{
				ctx: context.Background(),
				in:  input,
			},
			wantErr:    true,
			errMessage: ErrUnexpected.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectExec("INSERT INTO users").WithArgs(input.Name, input.Email, input.Password).WillReturnError(ErrUnexpected)
			},
		},
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				in:  input,
			},
			init: func(s sqlmock.Sqlmock) {
				s.ExpectExec("INSERT INTO users").WithArgs(input.Name, input.Email, input.Password).WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init(mockQuery)
			err := createUser(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("createUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("wrong error message got: %s, want: %s", err.Error(), tt.errMessage)
			}
			if err := mockQuery.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
