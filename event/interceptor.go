package event

import (
	"context"
	"strings"

	authorization "github.com/svladislav00-qq/event-microservices/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	userIDKey     contextKey = "user_id"
	roleKey       contextKey = "role"
	departmentKey contextKey = "department"
)

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata missing")
	}

	authHeader := md["authorization"]
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "token missing")
	}

	tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")

	claims, err := authorization.ParseJWT(tokenString)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	ctx = context.WithValue(ctx, userIDKey, claims.UserId)
	ctx = context.WithValue(ctx, roleKey, claims.Role)
	ctx = context.WithValue(ctx, departmentKey, claims.Department)

	return handler(ctx, req)
}
