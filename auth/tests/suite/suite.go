package suite

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/svladislav00-qq/event-microservices/auth/pb"
	"github.com/svladislav00-qq/event-microservices/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suit struct {
	*testing.T
	Cfg        *config.Config
	AuthClient pb.AuthServiceClient
}

const (
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suit) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoad()

	if cfg.TokenTTL == 0 {
		cfg.TokenTTL = time.Hour
	}
	ctx, cancelCtx := context.WithTimeout(context.Background(), 20*time.Second)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.NewClient(grpcAddress(&cfg), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal("grpc server connection is failed: %w", err)
	}
	return ctx, &Suit{
		T:          t,
		Cfg:        &cfg,
		AuthClient: pb.NewAuthServiceClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	port := cfg.GRPCPort
	if port == 0 {
		port = 44044
	}
	return net.JoinHostPort(grpcHost, strconv.Itoa(port))
}
