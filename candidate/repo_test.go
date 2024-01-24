package candidate

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
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

func TestSetCache(t *testing.T) {
	cache, _ := redismock.NewClientMock()
	type args struct {
		newCache *redis.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "empty cache",
			args:    args{},
			wantErr: ErrEmptyCache,
		},
		{
			name: "with mock cache",
			args: args{
				newCache: cache,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetCache(tt.args.newCache); err != tt.wantErr {
				t.Errorf("SetCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getCandidate(t *testing.T) {
	type args struct {
		ctx         context.Context
		excludedIDs []uint64
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
				ctx: context.Background(),
			},
			wantErr:    true,
			errMessage: ErrUnexpected.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT id, name, email, verified FROM users").WillReturnError(ErrUnexpected)
			},
		},
		{
			name: "no available user found",
			args: args{
				ctx: context.Background(),
			},
			wantErr:    true,
			errMessage: ErrNoCandidateAvailable.Error(),
			init: func(s sqlmock.Sqlmock) {
				s.ExpectQuery("SELECT id, name, email, verified FROM users").WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name: "success without exclusion",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
			wantUser: User{
				ID:       1,
				Name:     "Yoga",
				Email:    "yoga@mail.com",
				Verified: true,
			},
			init: func(s sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"id", "name", "email", "verified"}).AddRow(1, "Yoga", "yoga@mail.com", 1)
				s.ExpectQuery("SELECT id, name, email, verified FROM users").WillReturnRows(row)
			},
		},
		{
			name: "success with exclusion",
			args: args{
				ctx:         context.Background(),
				excludedIDs: []uint64{2, 4},
			},
			wantErr: false,
			wantUser: User{
				ID:       1,
				Name:     "Yoga",
				Email:    "yoga@mail.com",
				Verified: true,
			},
			init: func(s sqlmock.Sqlmock) {
				row := sqlmock.NewRows([]string{"id", "name", "email", "verified"}).AddRow(1, "Yoga", "yoga@mail.com", 1)
				s.ExpectQuery("SELECT id, name, email, verified FROM users").WillReturnRows(row)
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
			gotUser, err := getCandidate(tt.args.ctx, tt.args.excludedIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCandidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMessage {
				t.Errorf("wrong error message got: %s, want: %s", err.Error(), tt.errMessage)
			}
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("getCandidate() = %v, want %v", gotUser, tt.wantUser)
			}
			if err := mockQuery.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
