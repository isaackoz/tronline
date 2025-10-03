package cfg

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type envConfig struct {
	// debug, info, warn, error. Defaults to info
	LogLevel string `env:"LOG_LEVEL"`
	// if not empty or "", this is a production environment, otherwise it's development
	Production string `env:"ENVIRONMENT"`
	// Server address
	Addr string `env:"SERVER_ADDR" default:":8080"`
}

type Config struct {
	// the logging level. Defaults to info
	LogLevel slog.Level ``
	// if true, this is a production environment
	Production bool
	// Server address. Defaults to :8080
	Addr string
}

func LoadConfig(ctx context.Context) (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("load .env file", "error", err)
	}
	var c envConfig
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, fmt.Errorf("process envconfig: %w", err)
	}

	var cfg Config

	// log level
	switch strings.ToLower(c.LogLevel) {
	case "debug", "dbg", "d":
		cfg.LogLevel = slog.LevelDebug
	case "info", "information", "i":
		cfg.LogLevel = slog.LevelInfo
	case "warn", "warning", "w":
		cfg.LogLevel = slog.LevelWarn
	case "error", "err", "e":
		cfg.LogLevel = slog.LevelError
	default:
		cfg.LogLevel = slog.LevelInfo
	}

	// production
	if len(c.Production) > 0 {
		cfg.Production = true
	}

	// addr
	if len(c.Addr) > 0 {
		cfg.Addr = c.Addr
	} else {
		cfg.Addr = ":8080"
	}

	return &cfg, nil
}
