package authorization

import "context"

type userContextKey struct{}

func ContextsWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey{}).(*User)
	return user, ok
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value("token").(string)
	return token, ok
}
