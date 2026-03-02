package config

import (
	"log"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI              string        `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB               string        `env:"MONGO_DB" envDefault:"backend_challenge"`
	JWTSecret             string        `env:"JWT_SECRET" envDefault:"dev-secret"`
	HTTPPort              string        `env:"HTTP_PORT" envDefault:":8080"`
	GRPCPort              string        `env:"GRPC_PORT" envDefault:":50051"`
	LotteryReservationTTL time.Duration `env:"LOTTERY_RESERVATION_TTL" envDefault:"5m"`
	UserCountInterval     time.Duration `env:"USER_COUNT_INTERVAL" envDefault:"10s"`
}

func Load() Config {

	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found!")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	// Security Check: Ensure sensitive keys are changed from defaults if not in "dev"
	if cfg.JWTSecret == "dev-secret" {
		slog.Warn("JWT_SECRET is using default value! This is insecure for production.")
	}

	return cfg
}
