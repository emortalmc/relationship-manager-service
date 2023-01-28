package notifier

import (
	"context"
	"github.com/emortalmc/proto-specs/gen/go/model/relationship"
	"github.com/google/uuid"
)

type Notifier interface {
	FriendRequest(ctx context.Context, request *relationship.FriendRequest) error
	FriendAdded(ctx context.Context, senderId uuid.UUID, targetId uuid.UUID, senderUsername string) error
}
