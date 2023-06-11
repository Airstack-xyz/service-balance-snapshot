package repository

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/airstack-xyz/database-library/pkg/database"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"

	"github.com/airstack-xyz/lib/logger/testing/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestNewBalanceSnapshotRepository(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	balanceRepo := NewBalanceSnapshotRepository(db.DB, logger)
	assert.NotNil(t, balanceRepo)
	assert.Equal(t, db.DB, balanceRepo.db)
	assert.Equal(t, logger, balanceRepo.logger)
}

func TestCreateSnapshot(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return().Once()
	logger.On("Info", "mongoDB connection is closed!").Return().Once()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	balanceSnapshotRepo := NewBalanceSnapshotRepository(db.DB, logger)

	ctx := context.Background()

	t.Run("nil snapshot - err", func(t *testing.T) {
		err := balanceSnapshotRepo.CreateSnapshot(ctx, nil)
		assert.NotNil(t, err, "error should not be nil")
		assert.Equal(t, err, errors.New("inserting snapshot shouldn't be nil"), "err is not equal")
	})
	t.Run("should create snapshot", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()

		err := balanceSnapshotRepo.CreateSnapshot(ctx, &testBalanceSnapshotRecord)
		assert.Nil(t, err, "error should be nil")

		// clean up the test data
		balanceSnapshotCollection := db.DB.Database(constants.MONGO_TOKEN_DB).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})
}

func TestGetSnapshotByBlockNumber(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return().Once()
	logger.On("Info", "mongoDB connection is closed!").Return().Once()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	balanceSnapshotRepo := NewBalanceSnapshotRepository(db.DB, logger)
	balanceSnapshotCollection := db.DB.Database(constants.MONGO_TOKEN_DB).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)

	ctx := context.Background()

	t.Run("should return snapshot", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()

		if err := balanceSnapshotRepo.CreateSnapshot(context.Background(), &testBalanceSnapshotRecord); err != nil {
			assert.Fail(t, "error while creating snapshot test entry")
		}

		snapshot, err := balanceSnapshotRepo.GetSnapshotByBlockNumber(ctx, "1", testBalanceSnapshotRecord.Owner, testBalanceSnapshotRecord.TokenAddress, testBalanceSnapshotRecord.TokenId, 10000051)
		assert.Nil(t, err, "err should be nil")
		assert.NotNil(t, snapshot, "snapshot record shouldn't be nil")
		testBalanceSnapshotRecord.CreatedAt = snapshot.CreatedAt
		testBalanceSnapshotRecord.UpdatedAt = snapshot.UpdatedAt
		assert.Equal(t, *snapshot, testBalanceSnapshotRecord, "snapshot fetched is not expected one")

		// clean up the test data
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})

	t.Run("should not return snapshot", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()
		if err := balanceSnapshotRepo.CreateSnapshot(context.Background(), &testBalanceSnapshotRecord); err != nil {
			assert.Fail(t, "error while creating snapshot test entry")
		}

		snapshot, err := balanceSnapshotRepo.GetSnapshotByBlockNumber(ctx, "1", testBalanceSnapshotRecord.Owner, testBalanceSnapshotRecord.TokenAddress, testBalanceSnapshotRecord.TokenId, 10000053)
		assert.Nil(t, snapshot, "snapshot should be nil")
		assert.NotNil(t, err, " err shouldn't be nil")
		assert.Equal(t, err, mongo.ErrNoDocuments, "err should be mongo.ErrNoDocuments")

		// clean up the test data
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})
}

func TestFindFirstNearestHighSnapshotRecord(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return().Once()
	logger.On("Info", "mongoDB connection is closed!").Return().Once()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	balanceSnapshotRepo := NewBalanceSnapshotRepository(db.DB, logger)
	balanceSnapshotCollection := db.DB.Database(constants.MONGO_TOKEN_DB).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)

	ctx := context.Background()

	t.Run("should return first nearest snapshot with higher blockNo - success", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()
		if err := balanceSnapshotRepo.CreateSnapshot(context.Background(), &testBalanceSnapshotRecord); err != nil {
			assert.Fail(t, "error while creating snapshot test entry")
		}

		snapshot, err := balanceSnapshotRepo.FindFirstNearestHighSnapshotRecord(ctx, testBalanceSnapshotRecord.ChainID, testBalanceSnapshotRecord.Owner, testBalanceSnapshotRecord.TokenId, testBalanceSnapshotRecord.TokenAddress, 10000040)
		assert.Nil(t, err, "err should be nil")
		assert.NotNil(t, snapshot, "snapshot record shouldn't be nil")
		testBalanceSnapshotRecord.CreatedAt = snapshot.CreatedAt
		testBalanceSnapshotRecord.UpdatedAt = snapshot.UpdatedAt
		assert.Equal(t, *snapshot, testBalanceSnapshotRecord, "snapshot fetched is not expected one")

		// clean up the test data
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})
}

func TestUpdateSnapshotById(t *testing.T) {
	logger := new(mocks.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return().Once()
	logger.On("Info", "mongoDB connection is closed!").Return().Once()

	db := initDB(logger)
	defer func() {
		_ = db.Disconnect(context.Background())
	}()

	balanceSnapshotRepo := NewBalanceSnapshotRepository(db.DB, logger)
	balanceSnapshotCollection := db.DB.Database(constants.MONGO_TOKEN_DB).Collection(constants.BALANCE_SNAPSHOT_COLLECTION)

	ctx := context.Background()

	t.Run("invalid ID", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()
		if err := balanceSnapshotRepo.CreateSnapshot(context.Background(), &testBalanceSnapshotRecord); err != nil {
			assert.Fail(t, "error while creating snapshot test entry")
		}

		updateMap := make(map[string]interface{})
		expectedUpdatedEndBlockNo := uint64(10000060)
		expectedUpdatedEndBlockTimestamp := time.Date(2020, 05, 04, 13, 35, 58, 0, time.UTC)
		updateMap["endBlockNumber"] = expectedUpdatedEndBlockNo
		updateMap["endBlockTimestamp"] = expectedUpdatedEndBlockTimestamp

		noOfRecordsModified, err := balanceSnapshotRepo.UpdateSnapshotById(ctx, uuid.NewString(), updateMap)
		assert.NotNil(t, err, "err should not be nil")
		assert.Equal(t, 0, noOfRecordsModified, "no of updated snapshots object shouldn't be 0")
		assert.Equal(t, err, mongo.ErrNoDocuments, "err is not equal")

		// clean up the test data
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})

	t.Run("updateField map is nil", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()
		if err := balanceSnapshotRepo.CreateSnapshot(context.Background(), &testBalanceSnapshotRecord); err != nil {
			assert.Fail(t, "error while creating snapshot test entry")
		}

		noOfRecordsModified, err := balanceSnapshotRepo.UpdateSnapshotById(ctx, uuid.NewString(), nil)
		assert.NotNil(t, err, "err should not be nil")
		assert.Nil(t, noOfRecordsModified, "updated record should be nil")
		assert.Equal(t, 0, noOfRecordsModified, "no of updated snapshots object should be 0")
		assert.Equal(t, err, errors.New("updateFields map shouldn't be nil"), "err is not equal")

		// clean up the test data
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})

	t.Run("successfully updated", func(t *testing.T) {
		testBalanceSnapshotRecord := getTestBalanceSnapshotRecord()
		if err := balanceSnapshotRepo.CreateSnapshot(context.Background(), &testBalanceSnapshotRecord); err != nil {
			assert.Fail(t, "error while creating snapshot test entry")
		}

		updateMap := make(map[string]interface{})
		expectedUpdatedEndBlockNo := uint64(10000060)
		expectedUpdatedEndBlockTimestamp := time.Date(2020, 05, 04, 13, 35, 58, 0, time.UTC)
		updateMap["endBlockNumber"] = expectedUpdatedEndBlockNo
		updateMap["endBlockTimestamp"] = expectedUpdatedEndBlockTimestamp

		noOfRecordsModified, err := balanceSnapshotRepo.UpdateSnapshotById(ctx, testBalanceSnapshotRecord.ID, updateMap)
		assert.Nil(t, err)
		assert.Equal(t, 1, noOfRecordsModified, "no of updated snapshots object shouldn't be 0")

		// clean up the test data
		cleanUpErr := balanceSnapshotCollection.Drop(ctx)
		assert.Nil(t, cleanUpErr, "Error while cleaning up the test data")
	})
}

func getTestBalanceSnapshotRecord() model.BalanceSnapshot {
	return model.BalanceSnapshot{
		ID:                  "test-id",
		Blockchain:          "0xc40ae8a63dfa35694dbbbd1ba8b8d1ad72738a97",
		Owner:               "ethereum",
		ChainID:             "1",
		TokenAddress:        "0xdedd13be2d81a9afff28b4b45bd23ecfb9948dbc",
		TokenType:           constants.TOKEN_TYPE_ERC20,
		StartBlockNumber:    10000051,
		EndBlockNumber:      10000053,
		StartBlockTimestamp: time.Date(2020, 05, 04, 13, 35, 32, 0, time.UTC),
		EndBlockTimestamp:   time.Date(2020, 05, 04, 13, 35, 36, 0, time.UTC),
		Amount:              "7917500000000",
	}
}

func initDB(logger *mocks.ILogger) *database.Database {
	os.Setenv(constants.MONGODB_URI, "mongodb://localhost:27017")
	db := database.New(logger)
	if err := db.Connect(context.Background(), os.Getenv(constants.MONGODB_URI)); err != nil {
		log.Panic("unable to connect to mongodb!")
	}
	return db
}
