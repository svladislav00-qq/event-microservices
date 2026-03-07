package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
}

var AppConfig Config

func Load() {
	var err error = godotenv.Load()
	if err != nil {
		slog.Error("Warning: env file not found, using environment variables")
	}

	AppConfig = Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("Port"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
}
