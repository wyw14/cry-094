package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/cry-094/internal/bootstrap"
	"github.com/wyw14/cry-094/internal/config"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("configuration failed", zap.Error(err))
	}
	container, err := bootstrap.Build(bootstrap.Config{ParserBuild: cfg.ParserBuild, SigningKeyID: cfg.SigningKeyID, SigningSecret: cfg.SigningSecret, AuthSecret: cfg.AuthSecret})
	if err != nil {
		logger.Fatal("bootstrap failed", zap.Error(err))
	}
	server := &http.Server{Addr: cfg.Address, Handler: container.Router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", zap.String("address", cfg.Address))
		errCh <- server.ListenAndServe()
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case signal := <-signals:
		logger.Info("shutdown requested", zap.String("signal", signal.String()))
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", zap.Error(err))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
}
