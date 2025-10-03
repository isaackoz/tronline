package server

import (
	"log/slog"
	"os"

	"github.com/isaackoz/tronline/cfg"
	"github.com/lmittmann/tint"
)

func NewLogger(config *cfg.Config) *slog.Logger {
	var handler slog.Handler

	// color in the logs in development
	if !config.Production {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:  false,
			Level:      config.LogLevel,
			TimeFormat: "11:11:11",
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
			Level:     config.LogLevel,
		})
	}

	logger := slog.New(handler)

	return logger
}
