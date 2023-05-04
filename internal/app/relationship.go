package app

import (
	"context"
	"fmt"
	"github.com/emortalmc/proto-specs/gen/go/grpc/relationship"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"net"
	"relationship-manager-service/internal/config"
	"relationship-manager-service/internal/kafka"
	"relationship-manager-service/internal/repository"
	"relationship-manager-service/internal/service"
)

func Run(ctx context.Context, cfg *config.Config, logger *zap.SugaredLogger) {
	repo, err := repository.NewMongoRepository(ctx, cfg.MongoDB)
	if err != nil {
		logger.Fatalw("failed to create repository", "error", err)
	}

	notif := kafka.NewKafkaNotifier(cfg.Kafka, logger)
	if err != nil {
		logger.Fatalw("failed to create notifier", "error", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logger.Fatalw("failed to listen", "error", err)
	}

	s := grpc.NewServer(grpc.ChainUnaryInterceptor(
		grpczap.UnaryServerInterceptor(logger.Desugar(), grpczap.WithLevels(func(code codes.Code) zapcore.Level {
			if code != codes.Internal && code != codes.Unavailable && code != codes.Unknown {
				return zapcore.DebugLevel
			} else {
				return zapcore.ErrorLevel
			}
		})),
	))
	relationship.RegisterRelationshipServer(s, service.NewPermissionService(repo, logger, notif))
	logger.Infow("listening on port", "port", cfg.Port)

	err = s.Serve(lis)
	if err != nil {
		logger.Fatalw("failed to serve", "error", err)
		return
	}
}
