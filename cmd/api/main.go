package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"keeper.notifications.go/internal/config"
	"keeper.notifications.go/internal/notification"
	"keeper.notifications.go/internal/rabbitmq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()
	logger.Info("configuration loaded")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	notificationService, err := notification.NewService(logger, cfg)
	if err != nil {
		logger.Error("failed to initialize notification service", "error", err)
		os.Exit(1)
	}

	consumer := rabbitmq.NewConsumer(logger, cfg, notificationService)
	consumer.Start(ctx)

	logger.Info("notification service started. awaiting events.")

	<-ctx.Done()

	logger.Info("shutdown signal received, terminating service.")
}
