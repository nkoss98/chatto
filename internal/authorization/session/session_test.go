package session

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_jsonWebToken_GenerateTokens(t *testing.T) {
	const secret = "example_secret"
	const userID = "1"
	tests := []struct {
		name     string
		userID   string
		want     UserSession
		validate func(t *testing.T, u UserSession, userID string, secret []byte)
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:   "success - generate token",
			userID: userID,
			want:   UserSession{},
			validate: func(t *testing.T, u UserSession, userID string, secret []byte) {
				token, err := jwt.Parse(u.Token, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}

					return secret, nil
				})
				require.NoError(t, err)

				require.Equal(t, true, token.Valid)
				claims, ok := token.Claims.(jwt.MapClaims)
				require.Equal(t, true, ok)
				require.Equal(t, claims["id"], userID)
				ts, err := claims.GetExpirationTime()
				require.NoError(t, err)
				require.Equal(t, true, ok)
				require.WithinDuration(t, ts.UTC(), time.Now().Add(time.Hour*1), time.Minute*1)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := NewJsonWebToken(Config{
				TokenSecret: []byte(secret),
			})
			got, err := j.GenerateTokens(userID)
			tt.wantErr(t, err)
			tt.validate(t, got, tt.userID, j.config.TokenSecret)
		})
	}
}

func Test_jsonWebToken_ValidateToken(t *testing.T) {
	const secret = "wowmega"
	const userID = "1"

	type testArgs struct {
		userID string
	}
	scenario := func(ta testArgs) func(t *testing.T) {
		return func(t *testing.T) {
			j := jwtTokenManager{
				config: Config{
					TokenSecret: []byte(secret),
				},
			}
			got, err := j.GenerateTokens(userID)
			require.NoError(t, err)

			err = j.ValidateToken(got.Token)
			require.NoError(t, err)

			err = j.ValidateToken(got.RefreshToken)
			require.NoError(t, err)
		}
	}

	t.Run("success", scenario(testArgs{userID: userID}))
}
