package middlewares

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: handle more test cases
func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		prepareRequest func(string) *http.Request
	}{
		{
			name:       "success - not authorized",
			statusCode: 403,
			prepareRequest: func(u string) *http.Request {
				serverURL, err := url.Parse(u)
				require.NoError(t, err)
				r := &http.Request{
					Method: http.MethodGet,
					URL:    serverURL,
				}
				return r
			},
		},
		{
			name:       "403 - invalid token length",
			statusCode: 403,
			prepareRequest: func(s string) *http.Request {
				serverURL, err := url.Parse(s)
				require.NoError(t, err)
				r := &http.Request{
					Method: http.MethodGet,
					URL:    serverURL,
				}
				r.Header = make(http.Header)
				r.Header.Set("Authorization", "invalid")
				return r
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := httptest.NewServer(AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			})))

			r := tt.prepareRequest(s.URL)

			client := http.Client{}
			res, err := client.Do(r)
			require.Equal(t, res.StatusCode, tt.statusCode)
			require.NoError(t, err)
		})
	}
}
