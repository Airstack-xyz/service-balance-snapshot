package repository

import (
	"context"
	"time"

	"github.com/airstack-xyz/lib/logger"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ITokensRepository interface {
	GetToken(ctx context.Context, id string) (*model.Token, error)
}

type TokensRepository struct {
	db     *mongo.Client
	logger logger.ILogger
}

func NewTokensRepository(db *mongo.Client, logger logger.ILogger) *TokensRepository {
	return &TokensRepository{db: db, logger: logger}
}

func (r *TokensRepository) GetToken(ctx context.Context, id string) (*model.Token, error) {
	defer utils.RecordFunctionExecutionTime(ctx, "token.GetToken", r.logger)()
	childctx, cancel := context.WithTimeout(ctx, constants.CONTEXT_TIMEOUT_IN_SECONDS*time.Second)
	defer cancel()
	tokenCollection := r.db.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.TOKEN)
	var token model.Token
	filter := bson.M{"_id": id}
	if err := tokenCollection.FindOne(childctx, filter).Decode(&token); err != nil {
		r.logger.Errorf(ctx, "tokensList.FindTokens: error while finding tokens from mongo database error: %v", err)
		return nil, err
	}
	return &token, nil
}
