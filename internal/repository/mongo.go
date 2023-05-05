package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"relationship-manager-service/internal/config"
	"relationship-manager-service/internal/repository/model"
	"relationship-manager-service/internal/repository/registrytypes"
	"sync"
	"time"
)

type mongoRepository struct {
	Repository
	database    *mongo.Database
	friendColl  *mongo.Collection
	pFriendColl *mongo.Collection
	blockColl   *mongo.Collection
}

var (
	NotFriendsError      = errors.New("players are not friends")
	NoFriendRequestError = errors.New("no friend request found")

	AlreadyBlockedError = errors.New("player already blocked")
)

func NewMongoRepository(ctx context.Context, logger *zap.SugaredLogger, wg *sync.WaitGroup, cfg *config.MongoDBConfig) (Repository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI).SetRegistry(createCodecRegistry()))
	if err != nil {
		return nil, err
	}

	database := client.Database("relationship-manager")

	repo := &mongoRepository{
		database:    database,
		friendColl:  database.Collection("friend"),
		pFriendColl: database.Collection("pendingFriend"),
		blockColl:   database.Collection("block"),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		if err := client.Disconnect(ctx); err != nil {
			logger.Errorw("failed to disconnect from mongo", err)
		}
	}()

	return repo, nil
}

func (m *mongoRepository) CreateFriendConnection(ctx context.Context, conn model.FriendConnection) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.friendColl.InsertOne(ctx, conn)
	return err
}

func (m *mongoRepository) GetFriendConnections(ctx context.Context, playerId uuid.UUID) ([]*model.FriendConnection, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := m.friendColl.Find(ctx, bson.M{"$or": []interface{}{
		bson.M{"playerOneId": playerId},
		bson.M{"playerTwoId": playerId},
	}})
	if err != nil {
		return nil, err
	}

	var mongoResult []*model.FriendConnection
	err = cursor.All(ctx, &mongoResult)

	return mongoResult, err
}

func (m *mongoRepository) DeleteFriendConnection(ctx context.Context, playerId uuid.UUID, friendId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.friendColl.DeleteOne(ctx, bson.M{"$or": []interface{}{
		bson.M{"playerOneId": playerId, "playerTwoId": friendId},
		bson.M{"playerOneId": friendId, "playerTwoId": playerId},
	}})

	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return NotFriendsError
	}
	return nil
}

func (m *mongoRepository) AreFriends(ctx context.Context, playerOneId uuid.UUID, playerTwoId uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.friendColl.CountDocuments(ctx, bson.M{"$or": []bson.M{
		{"playerOneId": playerOneId, "playerTwoId": playerTwoId},
		{"playerOneId": playerTwoId, "playerTwoId": playerOneId},
	}})
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (m *mongoRepository) CreatePendingFriendConnection(ctx context.Context, conn model.PendingFriendConnection) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.pFriendColl.InsertOne(ctx, conn)
	// todo NOTE: already exists is mongo.IsDuplicateKeyError(err)
	return err
}

func (m *mongoRepository) DoesPendingFriendConnectionExist(ctx context.Context, requesterId uuid.UUID, targetId uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.pFriendColl.CountDocuments(ctx, bson.M{"requesterId": requesterId, "targetId": targetId})
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (m *mongoRepository) GetPendingFriendConnections(ctx context.Context, playerId uuid.UUID, opts DirectionOpts) ([]*model.PendingFriendConnection, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var orFilters = make([]bson.M, 0)
	if opts.Incoming {
		orFilters = append(orFilters, bson.M{"targetId": playerId})
	}
	if opts.Outgoing {
		orFilters = append(orFilters, bson.M{"requesterId": playerId})
	}

	if len(orFilters) == 0 {
		zap.S().Warnw("no direction options provided", "playerId", playerId)
		return nil, errors.New("no direction options provided")
	}

	cursor, err := m.pFriendColl.Find(ctx, bson.M{"$or": orFilters})
	if err != nil {
		return nil, err
	}

	var mongoResult []*model.PendingFriendConnection
	err = cursor.All(ctx, &mongoResult)
	return mongoResult, err
}

func (m *mongoRepository) DeletePendingFriendConnection(ctx context.Context, playerId uuid.UUID, friendId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.pFriendColl.DeleteOne(ctx, bson.M{"$or": []bson.M{
		{"requesterId": playerId, "targetId": friendId},
		{"requesterId": friendId, "targetId": playerId},
	}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return NoFriendRequestError
	}
	return nil
}

func (m *mongoRepository) DeletePendingFriendConnections(ctx context.Context, playerId uuid.UUID, opts DirectionOpts) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var orFilters = make([]bson.M, 0)
	if opts.Incoming {
		orFilters = append(orFilters, bson.M{"targetId": playerId})
	}
	if opts.Outgoing {
		orFilters = append(orFilters, bson.M{"requesterId": playerId})
	}

	if len(orFilters) == 0 {
		zap.S().Warnw("no direction options provided", "playerId", playerId)
		return 0, nil
	}

	result, err := m.pFriendColl.DeleteMany(ctx, bson.M{"$or": orFilters})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

func (m *mongoRepository) CreatePlayerBlock(ctx context.Context, block model.PlayerBlock) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.blockColl.InsertOne(ctx, block)

	if mongo.IsDuplicateKeyError(err) {
		return AlreadyBlockedError
	}

	return err
}

func (m *mongoRepository) DeletePlayerBlock(ctx context.Context, blockerId uuid.UUID, blockedId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.blockColl.DeleteOne(ctx, bson.M{"blockerId": blockerId, "blockedId": blockedId})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (m *mongoRepository) IsPlayerBlocked(ctx context.Context, playerOneId uuid.UUID, playerTwoId uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := m.blockColl.CountDocuments(ctx, bson.M{"$or": []bson.M{
		{"blockerId": playerOneId, "blockedId": playerTwoId},
		{"blockerId": playerTwoId, "blockedId": playerOneId},
	}})
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (m *mongoRepository) GetMutualBlocks(ctx context.Context, playerOneId uuid.UUID, playerTwoId uuid.UUID) ([]*model.PlayerBlock, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := m.blockColl.Find(ctx, bson.M{"$or": []bson.M{
		{"blockerId": playerOneId, "blockedId": playerTwoId},
		{"blockerId": playerTwoId, "blockedId": playerOneId},
	}})
	if err != nil {
		return nil, err
	}

	result := make([]*model.PlayerBlock, 0)
	err = cursor.All(ctx, &result)
	return result, err
}

func (m *mongoRepository) GetPlayerBlocks(ctx context.Context, playerId uuid.UUID) ([]*model.PlayerBlock, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := m.blockColl.Find(ctx, bson.M{"blockerId": playerId})
	if err != nil {
		return nil, err
	}
	result := make([]*model.PlayerBlock, 0)
	err = cursor.All(ctx, &result)
	return result, err
}

func createCodecRegistry() *bsoncodec.Registry {
	return bson.NewRegistryBuilder().
		RegisterTypeEncoder(registrytypes.UUIDType, bsoncodec.ValueEncoderFunc(registrytypes.UuidEncodeValue)).
		RegisterTypeDecoder(registrytypes.UUIDType, bsoncodec.ValueDecoderFunc(registrytypes.UuidDecodeValue)).
		Build()
}
