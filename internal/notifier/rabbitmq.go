package notifier

import (
	"context"
	"fmt"
	msgModel "github.com/emortalmc/proto-specs/gen/go/message/relationship"
	"github.com/emortalmc/proto-specs/gen/go/model/relationship"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"relationship-manager-service/internal/config"
)

const rabbitMqUriFormat = "amqp://%s:%s@%s:5672"

type rabbitMqNotifier struct {
	Notifier
	channel *amqp.Channel
}

func NewRabbitMqNotifier(cfg config.RabbitMQConfig) (Notifier, error) {
	conn, err := amqp.Dial(fmt.Sprintf(rabbitMqUriFormat, cfg.Username, cfg.Password, cfg.Host))
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &rabbitMqNotifier{
		channel: channel,
	}, nil
}

func (r *rabbitMqNotifier) FriendRequest(ctx context.Context, fReq *relationship.FriendRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5)
	defer cancel()

	msg := &msgModel.FriendRequestReceivedMessage{
		Request: fReq,
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	err = r.channel.PublishWithContext(ctx, "mc:proxy:all", "", false, false, amqp.Publishing{
		ContentType: "application/x-protobuf",
		Type:        string(msg.ProtoReflect().Descriptor().FullName()),
		Body:        bytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rabbitMqNotifier) FriendAdded(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID, senderUsername string) error {
	ctx, cancel := context.WithTimeout(ctx, 5)
	defer cancel()

	msg := &msgModel.FriendAddedMessage{
		SenderId:       senderId.String(),
		SenderUsername: senderUsername,
		RecipientId:    targetId.String(),
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	err = r.channel.PublishWithContext(ctx, "mc:proxy:all", "", false, false, amqp.Publishing{
		ContentType: "application/x-protobuf",
		Type:        string(msg.ProtoReflect().Descriptor().FullName()),
		Body:        bytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *rabbitMqNotifier) FriendRemoved(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5)
	defer cancel()

	msg := &msgModel.FriendRemovedMessage{
		SenderId:    senderId.String(),
		RecipientId: targetId.String(),
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	err = r.channel.PublishWithContext(ctx, "mc:proxy:all", "", false, false, amqp.Publishing{
		ContentType: "application/x-protobuf",
		Type:        string(msg.ProtoReflect().Descriptor().FullName()),
		Body:        bytes,
	})
	if err != nil {
		return err
	}
	return nil
}
