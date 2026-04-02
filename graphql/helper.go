package graphql

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func WithAuth(ctx context.Context) context.Context {
	token, _ := ctx.Value("token").(string)

	if token == "" {
		return ctx
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})

	return metadata.NewOutgoingContext(ctx, md)
}
