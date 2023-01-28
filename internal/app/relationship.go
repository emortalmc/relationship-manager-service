package app

import (
	"context"
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/grpc/relationship"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"relationship-manager-service/internal/config"
	"relationship-manager-service/internal/notifier"
	"relationship-manager-service/internal/repository"
	"relationship-manager-service/internal/service"
)

func Run(ctx context.Context, cfg *config.Config, logger *zap.SugaredLogger) {
	repo, err := repository.NewMongoRepository(ctx, cfg.MongoDB)
	if err != nil {
		logger.Fatalw("failed to create repository", "error", err)
	}

	notif, err := notifier.NewRabbitMqNotifier(cfg.RabbitMQ)
	if err != nil {
		logger.Fatalw("failed to create notifier", "error", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.Fatalw("failed to listen", "error", err)
	}

	s := grpc.NewServer()
	relationship.RegisterRelationshipServer(s, service.NewPermissionService(repo, notif))
	logger.Infow("listening on port", "port", cfg.Port)

	err = s.Serve(lis)
	if err != nil {
		logger.Fatalw("failed to serve", "error", err)
		return
	}
}
