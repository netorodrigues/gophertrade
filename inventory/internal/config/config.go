package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v3"
	"github.com/joho/godotenv"
)

type Config struct {
	Web struct {
		APIHost string `conf:"default:0.0.0.0:8081,env:INVENTORY_HTTP_PORT"`
	}
	GRPC struct {
		Port string `conf:"default:9091,env:INVENTORY_GRPC_PORT"`
	}
	DB struct {
		URL string `conf:"default:postgres://postgres:postgres@localhost:5432/inventory?sslmode=disable,env:DATABASE_URL"`
	}
	Kafka struct {
		Brokers string `conf:"default:localhost:9092,env:KAFKA_BROKERS"`
	}
	OTEL struct {
		ExporterEndpoint string `conf:"default:localhost:4317,env:OTEL_EXPORTER_OTLP_ENDPOINT"`
	}
}

func Load() (Config, error) {
	// Try to load .env from the root directory or current directory, ignore if not found
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	var cfg Config
	help, err := conf.Parse("", &cfg)

	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			os.Exit(0)
		}
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	// Post-process ports if they were given just as numbers
	if cfg.Web.APIHost != "" && cfg.Web.APIHost[0] != ':' {
		cfg.Web.APIHost = ":" + cfg.Web.APIHost
	}

	return cfg, nil
}
