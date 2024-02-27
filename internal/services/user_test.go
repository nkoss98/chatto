package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"scratch/api"
	"scratch/internal/authorization/session"
	db "scratch/internal/storage/database"
	mockdb "scratch/internal/storage/database/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAccountService_Login(t *testing.T) {
	customErr := errors.New("unexpected error")
	tests := []struct {
		name        string
		prepareMock func(t *testing.T, tokenMaker *session.MockIdentityGenerator, queries *mockdb.MockQuerier)
		want        api.LoginUserResponse
		error       error
	}{
		{
			name: "success - login user",
			prepareMock: func(t *testing.T, tokenMaker *session.MockIdentityGenerator, queries *mockdb.MockQuerier) {

				hash, err := bcrypt.GenerateFromPassword([]byte("Test123!"), bcrypt.MinCost)
				assert.NoError(t, err)

				queries.EXPECT().GetUserByEmail(gomock.Any(), "joedoe@gmail.com").
					Return(db.ScratchUser{
						ID:       1,
						Email:    "joedoe@gmail.com",
						Password: string(hash),
					}, nil)

				tokenMaker.EXPECT().GenerateTokens("1").Return(session.UserSession{
					RefreshToken: "refresh-token",
					Token:        "normal-token",
				}, nil)

			},
			want: api.LoginUserResponse{
				RefreshToken: "refresh-token",
				Token:        "normal-token",
			},
			error: nil,
		},
		{
			name: "fail - there is no user with that email",
			prepareMock: func(t *testing.T, tokenMaker *session.MockIdentityGenerator, queries *mockdb.MockQuerier) {
				queries.EXPECT().GetUserByEmail(gomock.Any(), "joedoe@gmail.com").
					Return(db.ScratchUser{}, sql.ErrNoRows)
			},
			want:  api.LoginUserResponse{},
			error: fmt.Errorf("login user: %w", UserNotFoundErr),
		},
		{
			name: "fail - random error from query",
			prepareMock: func(t *testing.T, tokenMaker *session.MockIdentityGenerator, queries *mockdb.MockQuerier) {
				queries.EXPECT().GetUserByEmail(gomock.Any(), "joedoe@gmail.com").
					Return(db.ScratchUser{}, customErr)
			},
			want:  api.LoginUserResponse{},
			error: fmt.Errorf("login: %w", customErr),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTokenMaker := session.NewMockIdentityGenerator(ctrl)
			mockQueries := mockdb.NewMockQuerier(ctrl)
			tt.prepareMock(t, mockTokenMaker, mockQueries)

			s := NewAccountService(mockQueries, mockTokenMaker, slog.Logger{})

			got, err := s.Login(context.Background(), api.LoginUserRequest{
				Email:    "joedoe@gmail.com",
				Password: "Test123!",
			})
			if err != nil {
				assert.EqualErrorf(t, err, tt.error.Error(), "Login() error = %v, wantErr %v", err, tt.error)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAccountService_CreateUser(t *testing.T) {

	tests := []struct {
		name        string
		prepareMock func(t *testing.T, queries *mockdb.MockQuerier)
		want        int
		wantErr     error
	}{
		{
			name: "success - create user",
			prepareMock: func(t *testing.T, queries *mockdb.MockQuerier) {
				queries.EXPECT().GetUserByEmail(gomock.Any(), "joedoe@gmail.com").Return(db.ScratchUser{}, sql.ErrNoRows)

				queries.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(db.ScratchUser{
					ID: 1,
				}, nil)
			},
			want:    1,
			wantErr: nil,
		},
		{
			name: "fail - user with that email already exist",
			prepareMock: func(t *testing.T, queries *mockdb.MockQuerier) {
				queries.EXPECT().GetUserByEmail(gomock.Any(), "joedoe@gmail.com").Return(db.ScratchUser{
					ID:       1,
					Name:     "",
					Email:    "joedoe@gmail.com",
					Password: "",
				}, nil)
			},
			want:    0,
			wantErr: UserExistErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockQueries := mockdb.NewMockQuerier(ctrl)
			tt.prepareMock(t, mockQueries)
			s := NewAccountService(mockQueries, nil, slog.Logger{})

			got, err := s.CreateUser(context.Background(), api.RegisterUserRequest{
				Email:    "joedoe@gmail.com",
				Name:     "konu33",
				Password: "Test123!",
			})

			if err != nil {
				assert.EqualErrorf(t, err, tt.wantErr.Error(),
					"CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
