package middlewares

import (
	"net/http"
	"scratch/internal/authorization/session"
	"strings"
)

// TODO: testy gdzie wrzucone fakowy struct niby ze validator ale nie

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			w.WriteHeader(403)
			return
		}
		token := strings.Split(h, " ")

		v := session.NewJsonWebToken(session.Config{
			TokenSecret: []byte("YELLOW SUBMARINE, BLACK WIZARDRY"),
		})

		if len(token) > 1 && token[1] != "" {
			err := v.ValidateToken(token[1])
			if err != nil {
				w.WriteHeader(403)
				return
			}
			next.ServeHTTP(w, r)
		}
		w.WriteHeader(403)
	})
}
