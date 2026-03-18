package main

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/svladislav00-qq/event-microservices/attendee"
	"github.com/svladislav00-qq/event-microservices/auth"
	"github.com/svladislav00-qq/event-microservices/pkg/config"
	"github.com/svladislav00-qq/event-microservices/pkg/logger"
	"github.com/tinrab/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Env             string
	DatabaseURL     string
	GRPCPort        int
	AuthGRPCAddress string
	JWTSecret       string
}

func main() {
	cfg := config.MustLoad()
	logg := logger.SetupLogger(cfg.Env)

	logg.Info("starting application", slog.String("env", cfg.Env), slog.Int("port", cfg.GRPCPort))

	var repo attendee.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err := attendee.NewPostgresRepository(logg, cfg.DatabaseURL)
		if err != nil {
			logg.Error("failed to connect to database", slog.Any("error", err))
			return err
		}

		repo = r
		return nil
	})
	defer repo.Close()
	logg.Info("database connected")

	authClient, err := auth.NewClient(cfg.AuthGRPCAddress)
	if err != nil {
		logg.Error("failed to create auth client", slog.Any("error", err))
		return
	}

	service := attendee.NewAttendeeService(
		logg,
		repo,
		authClient,
	)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(attendee.AuthInterceptor(cfg.JWTSecret)),
	)
	attendee.NewServer(grpcServer, service)

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
