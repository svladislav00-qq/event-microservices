package main

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/svladislav00-qq/event-microservices/auth"
	"github.com/svladislav00-qq/event-microservices/pkg/config"
	"github.com/svladislav00-qq/event-microservices/pkg/logger"
	"github.com/tinrab/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Env         string        `envconfig:"ENV" default:"local"`
	DatabaseURL string        `envconfig:"DATABASE_URL"`
	GRPCPort    int           `envconfig:"GRPC_PORT" default:"8080"`
	TokenTTL    time.Duration `envconfig:"TOKEN_TTL" default:"1h"`
}

func main() {
	cfg := config.MustLoad()
	logg := logger.SetupLogger(cfg.Env)

	logg.Info("starting application", slog.String("env", cfg.Env), slog.Int("port", cfg.GRPCPort))

	var repo auth.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repo, err = auth.NewPostgresRepository(logg, cfg.DatabaseURL)
		if err != nil {
			logg.Error("failed to connect to database", slog.Any("error", err))
			return err
		}
		return nil
	})
	defer repo.Close()
	logg.Info("database connected")

	service := auth.New(logg, repo, repo, time.Hour)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.AuthInterceptor),
	)
	auth.NewAuthServer(grpcServer, service)

	logg.Info("AUTH VERSION: NEW BUILD WITH GetUsersByIDs")
	logg.Info("starting application", slog.Int("port", cfg.GRPCPort))

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logg.Error("failed to listen", slog.Any("error", err))
		return
	}

	logg.Info("starting grpc server", slog.Int("port", cfg.GRPCPort))
	if err := grpcServer.Serve(lis); err != nil {
		logg.Error("server stopped", slog.Any("error", err))
	}

}
