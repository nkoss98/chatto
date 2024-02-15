package middlewares

import (
	"context"
	"net/http"
)

func customMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := context.WithValue(request.Context(), "services", "poka cyce")

		next.ServeHTTP(writer, request.WithContext(ctx))
	})

}
