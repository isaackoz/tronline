package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/isaackoz/tronline/cfg"
)

// ref: https://victoriametrics.com/blog/go-graceful-shutdown/#summary
const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownPeriodHard  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

var isShuttingDown atomic.Bool

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := cfg.LoadConfig(ctx)
	if err != nil {
		log.Fatal("load config", err)
	}

	logger := NewLogger(config)
	slog.SetDefault(logger)

	slog.Info("starting server",
		"version", "v0.1.0",
		"production", config.Production,
		"log_level", config.LogLevel.String(),
	)

	mux := http.NewServeMux()

	// readiness
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if isShuttingDown.Load() {
			http.Error(w, "Shutting down", http.StatusServiceUnavailable)
		}
		fmt.Fprintln(w, "OK")
	})

	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: config.Addr,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}
	go func() {
		slog.Info("listening", "addr", config.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-ctx.Done()
	stop()
	isShuttingDown.Store(true)
	slog.Info("shutting down...")
	time.Sleep(_readinessDrainDelay)
	slog.Info("drain delay passed, shutting down connections gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()
	err = server.Shutdown(shutdownCtx)
	stopOngoingGracefully()
	if err != nil {
		slog.Error("graceful shutdown did not complete in time", "err", err)
		time.Sleep(_shutdownPeriodHard)
	}
	slog.Info("shutdown complete. goodbye")
}
