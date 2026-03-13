package main

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/svladislav00-qq/event-microservices/event"
	"github.com/svladislav00-qq/event-microservices/pkg/config"
	"github.com/svladislav00-qq/event-microservices/pkg/logger"
	"github.com/svladislav00-qq/event-microservices/pkg/minio"
	"github.com/tinrab/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Env         string
	DatabaseURL string
}

func main() {
	cfg := config.MustLoad()
	logg := logger.SetupLogger(cfg.Env)

	logg.Info("starting application", slog.String("env", cfg.Env), slog.Int("port", cfg.GRPCPort))

	var repo event.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err := event.NewPostgresRepository(logg, cfg.DatabaseURL)
		if err != nil {
			logg.Error("failed to connect to database", slog.Any("error", err))
			return err
		}

		repo = r
		return nil
	})
	defer repo.Close()
	logg.Info("database connected")

	minio.CreateCloud()

	storage := event.New(
		minio.MinioClient,
		minio.MinioBucket,
		minio.MinioPublicURL,
		logg,
	)

	service := event.NewEventService(
		logg,
		repo,
		repo,
		storage,
	)

	logg.Info("minio connected")

	grpcServer := grpc.NewServer()
	event.NewEventServer(grpcServer, service)

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
