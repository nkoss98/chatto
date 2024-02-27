package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"scratch/api"
	"scratch/internal/authorization/session"
	"scratch/internal/services"
	storage "scratch/internal/storage/database"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type response struct {
	Id int `json:"id"`
}

func Test_accountHandler_PostRegister(t *testing.T) {
	type testArgs struct {
		prepareDB      func(t *testing.T)
		verifyResponse func(t *testing.T, r *http.Response)
	}
	scenario := func(tt testArgs) func(t *testing.T) {
		return func(t *testing.T) {

			srv := initService(t)
			client := srv.Client()
			tt.prepareDB(t)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
				fmt.Sprintf("%v/register", srv.URL), strings.NewReader(`{
   			 	 "email" :"norbi12@wp.pl",
    			 "name":"Konu33",
    			 "password":"Test123!"}`))
			assert.NoError(t, err)

			res, err := client.Do(req)
			assert.NoError(t, err)

			tt.verifyResponse(t, res)

			defer res.Body.Close()

		}
	}

	t.Run("success - register as user", scenario(testArgs{
		prepareDB: func(t *testing.T) {
			// nothing to do
		},
		verifyResponse: func(t *testing.T, r *http.Response) {
			wanted := response{Id: 1}
			t.Helper()
			assert.Equal(t, http.StatusCreated, r.StatusCode)
			var res response
			err := json.NewDecoder(r.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, wanted, res)

		},
	}))

	t.Run("user with that email already exist", scenario(testArgs{
		prepareDB: func(t *testing.T) {
			t.Helper()
			tokenMaker := session.NewJsonWebToken(session.Config{
				TokenSecret: []byte("real secret"),
			})
			ctrl := services.NewAccountService(storage.New(db), tokenMaker, slog.Logger{})

			_, err := ctrl.CreateUser(context.Background(), api.RegisterUserRequest{
				Email:    "norbi1@wp.pl",
				Name:     "konu33",
				Password: "Test123!",
			})
			assert.NoError(t, err)
		},
		verifyResponse: func(t *testing.T, r *http.Response) {
			wanted := api.ErrorResponse{Error: "user with that email already exists"}
			t.Helper()
			assert.Equal(t, http.StatusConflict, r.StatusCode)
			var res api.ErrorResponse
			err := json.NewDecoder(r.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, wanted, res)
		},
	}))
}

func Test_accountHandler_PostLogin(t *testing.T) {

	tests := []struct {
		name               string
		requestBody        string
		wantResponseStatus int
		prepareDB          func(t *testing.T)
		verifyResponse     func(t *testing.T, r *http.Response)
	}{
		{
			name:               "success login",
			requestBody:        `{"email":"norbi22@wp.pl", "password":"Test123!"}`,
			wantResponseStatus: 200,
			prepareDB: func(t *testing.T) {
				t.Helper()
				tokenMaker := session.NewJsonWebToken(session.Config{
					TokenSecret: []byte("real secret"),
				})
				ctrl := services.NewAccountService(storage.New(db), tokenMaker, slog.Logger{})

				_, err := ctrl.CreateUser(context.Background(), api.RegisterUserRequest{
					Email:    "norbi22@wp.pl",
					Name:     "konu33",
					Password: "Test123!",
				})
				assert.NoError(t, err)
			},
			verifyResponse: func(t *testing.T, r *http.Response) {
				var resBody api.LoginUserResponse
				err := json.NewDecoder(r.Body).Decode(&resBody)
				assert.NoError(t, err)

				assert.NotEmpty(t, resBody)

			},
		},
		{
			name:               "no user with that email - 404",
			requestBody:        `{"email":"norbi33@wp.pl", "password":"Test123!"}`,
			wantResponseStatus: http.StatusNotFound,
			prepareDB: func(t *testing.T) {
				t.Helper()
			},
			verifyResponse: func(t *testing.T, r *http.Response) {
				var resBody api.ErrorResponse
				err := json.NewDecoder(r.Body).Decode(&resBody)
				assert.NoError(t, err)

				assert.NotEmpty(t, resBody)
			},
		},
		{
			name:               "invalid password - 400",
			requestBody:        `{"email":"norbi@wp.pl", "password":"invalid!"}`,
			wantResponseStatus: http.StatusBadRequest,
			prepareDB: func(t *testing.T) {
				t.Helper()
				tokenMaker := session.NewJsonWebToken(session.Config{
					TokenSecret: []byte("real secret"),
				})
				ctrl := services.NewAccountService(storage.New(db), tokenMaker, slog.Logger{})

				_, err := ctrl.CreateUser(context.Background(), api.RegisterUserRequest{
					Email:    "norbi@wp.pl",
					Name:     "konu33",
					Password: "Test123!",
				})
				assert.NoError(t, err)
			},
			verifyResponse: func(t *testing.T, r *http.Response) {
				var resBody api.ErrorResponse
				err := json.NewDecoder(r.Body).Decode(&resBody)
				assert.NoError(t, err)

				assert.NotEmpty(t, resBody)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := initService(t)
			client := srv.Client()
			tt.prepareDB(t)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
				fmt.Sprintf("%v/login", srv.URL), strings.NewReader(tt.requestBody))
			assert.NoError(t, err)

			res, err := client.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantResponseStatus, res.StatusCode)

			tt.verifyResponse(t, res)

			defer res.Body.Close()
		})
	}
}
