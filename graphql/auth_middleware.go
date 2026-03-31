package graphql

import (
	"context"
	"net/http"
	"strings"

	pkgauth "github.com/svladislav00-qq/event-microservices/pkg/auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("authorization")

		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := pkgauth.ParseJWT(token)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"errors": [{"message":"invalid token"}]}`))
			return
		}

		user := &pkgauth.User{
			ID:         claims.ID,
			Username:   claims.Username,
			Role:       claims.Role,
			Department: claims.Department,
		}
		ctx := pkgauth.ContextsWithUser(r.Context(), user)
		ctx = context.WithValue(ctx, "token", token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
