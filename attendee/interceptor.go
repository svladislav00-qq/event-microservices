package attendee

import (
	"context"
	"strings"

	authorization "github.com/svladislav00-qq/event-microservices/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type userIDKeyType struct{}
type roleKeyType struct{}

var (
	userIDKey = userIDKeyType{}
	roleKey   = roleKeyType{}
)

const RoleModerator = "moderator"

func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "no metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "no auth header")
		}

		auth := authHeader[0]

		if !strings.HasPrefix(auth, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid auth scheme")
		}

		token := strings.TrimPrefix(auth, "Bearer ")

		claims, err := authorization.ParseJWT(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if claims.Role == "" {
			return nil, status.Error(codes.Unauthenticated, "no role in token")
		}

		if isModeratorOnlyMethod(info.FullMethod) && claims.Role != RoleModerator {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		ctx = context.WithValue(ctx, userIDKey, claims.UserId)
		ctx = context.WithValue(ctx, roleKey, claims.Role)

		return handler(ctx, req)
	}
}

func isModeratorOnlyMethod(method string) bool {
	switch method {
	case "/pb.AttendeeService/GetEventAttendees",
		"/pb.AttendeeService/ExportAttendeesTable":
		return true
	default:
		return false
	}
}
