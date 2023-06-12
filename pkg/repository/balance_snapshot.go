package repository

import (
	"context"
	"errors"
	"time"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/utils"

	helper "github.com/airstack-xyz/database-library/pkg/utils"
	"github.com/airstack-xyz/lib/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BalanceSnapshotRepository struct {
	db     *mongo.Client
	logger logger.ILogger
}

type IBalanceSnapshotRepository interface {
	GetSnapshotByBlockNumber(ctx context.Context, chainId string, owner string, tokenAddress string, tokenId string, blockNumber uint) (*model.BalanceSnapshot, error)
	CreateSnapshot(ctx context.Context, snapshot *model.BalanceSnapshot) error
	FindFirstNearestHighSnapshotRecord(ctx context.Context, chainId string, owner string, tokenId string, tokenAddress string, blockNumber uint) (*model.BalanceSnapshot, error)
	UpdateSnapshotById(ctx context.Context, id string, updateFields map[string]interface{}) (int64, error)
	BulkWriteSnapshot(ctx context.Context, writeModels []mongo.WriteModel) error
}

func NewBalanceSnapshotRepository(db *mongo.Client, logger logger.ILogger) *BalanceSnapshotRepository {
	return &BalanceSnapshotRepository{db: db, logger: logger}
}

func (r *BalanceSnapshotRepository) GetSnapshotByBlockNumber(ctx context.Context, chainId string, owner string, tokenAddress string, tokenId string, blockNumber uint) (*model.BalanceSnapshot, error) {
	defer utils.RecordFunctionExecutionTime(ctx, "GetSnapshotByBlockNumber", r.logger)()
	childctx, cancel := context.WithTimeout(ctx, constants.CONTEXT_TIMEOUT_IN_SECONDS*time.Second)
	defer cancel()
	balanceSnapshotCollection := r.db.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)
	filterArray := bson.A{
		bson.D{{Key: "startBlockNumber", Value: bson.D{{Key: "$lte", Value: blockNumber}}}},
		bson.D{{Key: "endBlockNumber", Value: bson.D{{Key: "$gt", Value: blockNumber}}}},
		bson.D{{Key: "chainId", Value: chainId}},
		bson.D{{Key: "tokenAddress", Value: tokenAddress}},
		bson.D{{Key: "owner", Value: owner}},
	}
	if tokenId != "" {
		filterArray = append(filterArray, bson.D{{Key: "tokenId", Value: tokenId}})
	}
	filter := bson.D{
		{Key: "$and", Value: filterArray}}
	var result *model.BalanceSnapshot
	if err := balanceSnapshotCollection.FindOne(childctx, filter).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *BalanceSnapshotRepository) CreateSnapshot(ctx context.Context, snapshot *model.BalanceSnapshot) error {
	defer utils.RecordFunctionExecutionTime(ctx, "CreateSnapshot", r.logger)()
	if snapshot == nil {
		return errors.New("inserting snapshot shouldn't be nil")
	}

	childctx, cancel := context.WithTimeout(ctx, constants.CONTEXT_TIMEOUT_IN_SECONDS*time.Second)
	defer cancel()
	balanceSnapshotCollection := r.db.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)

	snapshot.CreatedAt = time.Now().UTC()
	snapshot.UpdatedAt = time.Now().UTC()
	_, err := balanceSnapshotCollection.InsertOne(childctx, snapshot)

	if err != nil {
		r.logger.Error("BalanceSnapshotRepository.CreateSnapshot: error while creating balanceSnapshot - %v", err.Error())
	}
	return err
}

func (r *BalanceSnapshotRepository) BulkWriteSnapshot(ctx context.Context, writeModels []mongo.WriteModel) error {
	defer utils.RecordFunctionExecutionTime(ctx, "BulkWriteSnapshot", r.logger)()
	if len(writeModels) == 0 {
		r.logger.Info("balance snapshot: BulkWriteSnapshot :: write model is empty")
		return nil
	}

	childctx, cancel := context.WithTimeout(ctx, constants.CONTEXT_TIMEOUT_IN_SECONDS*time.Second)
	defer cancel()
	balanceSnapshotCollection := r.db.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)

	opts := options.BulkWrite()
	opts.SetOrdered(true)

	if _, err := balanceSnapshotCollection.BulkWrite(childctx, writeModels, opts); err != nil {
		return err
	}

	return nil
}

func (r *BalanceSnapshotRepository) FindFirstNearestHighSnapshotRecord(ctx context.Context, chainId string, owner string, tokenId string, tokenAddress string, blockNumber uint) (*model.BalanceSnapshot, error) {
	defer utils.RecordFunctionExecutionTime(ctx, "FindFirstNearestHighSnapshotRecord", r.logger)()
	childctx, cancel := context.WithTimeout(ctx, constants.CONTEXT_TIMEOUT_IN_SECONDS*time.Second)
	defer cancel()
	balanceSnapshotCollection := r.db.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)
	filterArray := bson.A{
		bson.D{{Key: "chainId", Value: chainId}},
		bson.D{{Key: "tokenAddress", Value: tokenAddress}},
		bson.D{{Key: "owner", Value: owner}},
		bson.D{{Key: "startBlockNumber", Value: bson.D{{Key: "$gt", Value: blockNumber}}}},
	}
	if tokenId != "" {
		filterArray = append(filterArray, bson.D{{Key: "tokenId", Value: tokenId}})
	}
	filter := bson.D{
		{Key: "$and", Value: filterArray}}
	var result *model.BalanceSnapshot
	if err := balanceSnapshotCollection.FindOne(childctx, filter, options.FindOne().SetSort(bson.D{{Key: "startBlockNumber", Value: 1}})).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *BalanceSnapshotRepository) UpdateSnapshotById(ctx context.Context, id string, updateFields map[string]interface{}) (int64, error) {
	defer utils.RecordFunctionExecutionTime(ctx, "UpdateSnapshotById", r.logger)()
	if updateFields == nil {
		return 0, errors.New("updateFields map shouldn't be nil")
	}
	childctx, cancel := context.WithTimeout(ctx, constants.CONTEXT_TIMEOUT_IN_SECONDS*time.Second)
	defer cancel()
	balanceSnapshotCollection := r.db.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)
	updateFields["updatedAt"] = time.Now()
	updateFieldsBson := helper.ConvertStructToBsonM(updateFields)
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: updateFieldsBson},
	}
	updateResult, err := balanceSnapshotCollection.UpdateOne(childctx, filter, update)
	if err != nil {
		return 0, err
	}
	return updateResult.ModifiedCount, nil
}
