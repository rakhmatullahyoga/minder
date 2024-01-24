package auth

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
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
	type args struct {
		ctx context.Context
		in  Registration
	}
	input := Registration{Name: "Yoga", Password: "password", Email: "yoga@mail.com"}
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

	mockDb, mockQuery, _ := sqlmock.New()
	defer mockDb.Close()
	testDb := sqlx.NewDb(mockDb, "sqlmock")
	_ = SetDB(testDb)
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

func Test_isUserEmailExisted(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name        string
		args        args
		wantExisted bool
		wantErr     bool
		errMessage  string
		init        func(sqlmock.Sqlmock)
	}{
		{
			name: "error from database",
			args: args{
				ctx:   context.Background(),
				email: "yoga@mail.com",
			},
			wantExisted: false,
			wantErr:     true,
			errMessage:  ErrUnexpected.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT EXISTS").WithArgs("yoga@mail.com").WillReturnError(ErrUnexpected)
			},
		},
		{
			name: "success return exist",
			args: args{
				ctx:   context.Background(),
				email: "yoga@mail.com",
			},
			wantExisted: true,
			wantErr:     false,
			init: func(s sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"result"}).AddRow(1)
				s.ExpectQuery("SELECT EXISTS").WithArgs("yoga@mail.com").WillReturnRows(row)
			},
		},
		{
			name: "success return not exist",
			args: args{
				ctx:   context.Background(),
				email: "yoga@mail.com",
			},
			wantExisted: false,
			wantErr:     false,
			init: func(s sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"result"}).AddRow(0)
				s.ExpectQuery("SELECT EXISTS").WithArgs("yoga@mail.com").WillReturnRows(row)
			},
		},
	}

	mockDb, mockQuery, _ := sqlmock.New()
	defer mockDb.Close()
	testDb := sqlx.NewDb(mockDb, "sqlmock")
	_ = SetDB(testDb)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init(mockQuery)
			gotExisted, err := isUserEmailExisted(tt.args.ctx, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("isUserEmailExisted() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("wrong error message got: %s, want: %s", err.Error(), tt.errMessage)
			}
			if gotExisted != tt.wantExisted {
				t.Errorf("isUserEmailExisted() = %v, want %v", gotExisted, tt.wantExisted)
			}
			if err := mockQuery.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_getUserByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name       string
		args       args
		wantUser   User
		wantErr    bool
		errMessage string
		init       func(sqlmock.Sqlmock)
	}{
		{
			name: "unexpected error from database",
			args: args{
				ctx:   context.Background(),
				email: "yoga@mail.com",
			},
			wantErr:    true,
			errMessage: ErrUnexpected.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT id, name, email, password, verified FROM users").WithArgs("yoga@mail.com").WillReturnError(ErrUnexpected)
			},
		},
		{
			name: "record not exists",
			args: args{
				ctx:   context.Background(),
				email: "yoga@mail.com",
			},
			wantErr:    true,
			errMessage: ErrUserNotFound.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT id, name, email, password, verified FROM users").WithArgs("yoga@mail.com").WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name: "success",
			args: args{
				ctx:   context.Background(),
				email: "yoga@mail.com",
			},
			wantUser: User{
				ID:       1,
				Name:     "Yoga",
				Email:    "yoga@mail.com",
				Password: "password",
				Verified: true,
			},
			wantErr: false,
			init: func(s sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"id", "name", "email", "password", "verified"}).AddRow(1, "Yoga", "yoga@mail.com", "password", 1)
				s.ExpectQuery("SELECT id, name, email, password, verified FROM users").WithArgs("yoga@mail.com").WillReturnRows(row)
			},
		},
	}

	mockDb, mockQuery, _ := sqlmock.New()
	defer mockDb.Close()
	testDb := sqlx.NewDb(mockDb, "sqlmock")
	_ = SetDB(testDb)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init(mockQuery)
			gotUser, err := getUserByEmail(tt.args.ctx, tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("getUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("wrong error message got: %s, want: %s", err.Error(), tt.errMessage)
			}
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("getUserByEmail() = %v, want %v", gotUser, tt.wantUser)
			}
			if err := mockQuery.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
