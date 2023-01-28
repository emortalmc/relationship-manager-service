package main

import (
	"context"
	"go.uber.org/zap"
	"log"
	"relationship-manager-service/internal/app"
	"relationship-manager-service/internal/config"
)

func main() {
	logger, err := createLogger()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		logger.Fatalw("failed to load config", "error", err)
	}

	ctx := context.Background()

	app.Run(ctx, cfg, logger)
}

func createLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
