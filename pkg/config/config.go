package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env             string
	DatabaseURL     string
	Port            int
	JWTSecret       string
	GRPCPort        int
	AuthGRPCAddress string
}

func MustLoad() Config {

	err := godotenv.Load("../../.env")
	if err != nil {
		slog.Warn("env file not found")
	}
	grpcPort, _ := strconv.Atoi(os.Getenv("GRPC_PORT"))
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	return Config{
		GRPCPort:        grpcPort,
		Env:             os.Getenv("ENV"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Port:            port,
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AuthGRPCAddress: os.Getenv("AUTH_GRPC_ADDRESS"),
	}
}
