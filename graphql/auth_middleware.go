package main

import (
	"net/http"
	"strings"

	"github.com/svladislav00-qq/event-microservices/auth"
	pkgauth "github.com/svladislav00-qq/event-microservices/pkg/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := auth.ParseJWT(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		user := &pkgauth.User{
			ID:         claims.ID,
			Username:   claims.Username,
			Role:       claims.Role,
			Department: claims.Department,
		}
		ctx := pkgauth.ContextsWithUser(r.Context(), user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
