package repository

import (
	"context"
	"testing"
	"time"

	"github.com/airstack-xyz/lib/logger/testing/mocks"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestNewTokenRepository(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	tokenRepo := NewTokensRepository(db.DB, logger)
	assert.NotNil(t, tokenRepo)
	assert.Equal(t, db.DB, tokenRepo.db)
	assert.Equal(t, logger, tokenRepo.logger)
}

func TestGetToken(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Return()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	tokenRepo := NewTokensRepository(db.DB, logger)

	t.Setenv(constants.CHAINID, "1")

	t.Run("token id empty", func(t *testing.T) {
		token, err := tokenRepo.GetToken(context.Background(), "")
		assert.Nil(t, token, "token should be nil")
		assert.NotNil(t, err, "error shouldn't be nil")
		assert.EqualError(t, err, "token id can't be empty")
	})

	t.Run("should return err no doc err when invalid id is given", func(t *testing.T) {
		token, err := tokenRepo.GetToken(context.Background(), "random-id")
		assert.Nil(t, token, "token should be nil")
		assert.NotNil(t, err, "error shouldn't be nil")
		assert.Equal(t, err, mongo.ErrNoDocuments)
	})

	t.Run("should return token if valid id is given", func(t *testing.T) {
		sampleToken := getSampleTokenRecord()
		err := tokenRepo.CreateToken(context.Background(), &sampleToken)
		assert.Nil(t, err, "couldn't create token")
		token, err := tokenRepo.GetToken(context.Background(), sampleToken.ID)
		assert.Nil(t, err, "getToken error should be nil")
		assert.NotNil(t, token, "token shouldn't be nil")
		assert.Equal(t, sampleToken.ID, token.ID)

		// clean up the test data
		db := initDB(logger)
		defer func() {
			_ = db.Disconnect(context.Background())
		}()

		balanceSnapshotCollection := db.DB.Database(getDB(constants.MONGO_TOKEN_DB)).Collection(constants.TOKEN)
		cleanUpErr := balanceSnapshotCollection.Drop(context.Background())
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})
}

func getSampleTokenRecord() model.Token {
	return model.Token{
		ID:                    "10x1130547436810db920fa73681c946fea15e9b758",
		Address:               "0x1130547436810db920fa73681c946fea15e9b758",
		ChainId:               "1",
		Decimals:              utils.Ptr(uint64(8)),
		LastTransferHash:      "2187564e3c02218157f9bac25511a130bdfb37c9d19415209f8c957f884499fb",
		LastTransferTimestamp: time.Date(2017, 10, 11, 12, 07, 49, 0, time.UTC),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Logo:                  model.LogoSizes{},
		TransferCount:         24,
		Type:                  "ERC20",
		Blockchain:            "ethereum",
		CurrentHolderCount:    31,
	}
}
