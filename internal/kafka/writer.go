package kafka

import (
	"context"
	"fmt"
	pbmsg "github.com/emortalmc/proto-specs/gen/go/message/relationship"
	"github.com/emortalmc/proto-specs/gen/go/model/relationship"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"relationship-manager-service/internal/config"
	"time"
)

const writeTopic = "relationship-manager"

type Notifier interface {
	FriendRequest(ctx context.Context, request *relationship.FriendRequest) error
	FriendAdded(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID, senderUsername string) error
	FriendRemoved(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID) error
}

type kafkaNotifier struct {
	w *kafka.Writer
}

func NewKafkaNotifier(cfg *config.KafkaConfig, logger *zap.SugaredLogger) Notifier {
	w := &kafka.Writer{
		Addr:         kafka.TCP(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		Topic:        writeTopic,
		Balancer:     &kafka.LeastBytes{},
		Async:        true,
		BatchTimeout: 100 * time.Millisecond,
		ErrorLogger:  kafka.LoggerFunc(logger.Errorw),
	}

	return &kafkaNotifier{w: w}
}

func (k *kafkaNotifier) FriendRequest(ctx context.Context, fReq *relationship.FriendRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg := &pbmsg.FriendRequestReceivedMessage{
		Request: fReq,
	}

	if err := k.writeMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %s", err)
	}

	return nil
}

func (k *kafkaNotifier) FriendAdded(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID, senderUsername string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg := &pbmsg.FriendAddedMessage{
		SenderId:       senderId.String(),
		SenderUsername: senderUsername,
		RecipientId:    targetId.String(),
	}

	if err := k.writeMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %s", err)
	}

	return nil
}

func (k *kafkaNotifier) FriendRemoved(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	msg := &pbmsg.FriendRemovedMessage{
		SenderId:    senderId.String(),
		RecipientId: targetId.String(),
	}

	if err := k.writeMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %s", err)
	}

	return nil
}

func (k *kafkaNotifier) writeMessage(ctx context.Context, msg proto.Message) error {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal proto to bytes: %s", err)
	}

	return k.w.WriteMessages(ctx, kafka.Message{
		Headers: []kafka.Header{{Key: "X-Proto-Type", Value: []byte(msg.ProtoReflect().Descriptor().FullName())}},
		Value:   bytes,
	})
}
