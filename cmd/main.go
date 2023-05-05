package main

import (
	"go.uber.org/zap"
	"log"
	"relationship-manager-service/internal/app"
	"relationship-manager-service/internal/config"
)

func main() {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := createLogger(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(cfg, logger)
}

func createLogger(cfg *config.Config) (*zap.SugaredLogger, error) {
	var logger *zap.Logger
	var err error
	if cfg.Development {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
