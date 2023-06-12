package service

import (
	"context"
	"testing"

	cacheMock "github.com/airstack-xyz/lib/cache/testing/mocks"
	distributedlock "github.com/airstack-xyz/lib/distributed-lock"
	loggerMock "github.com/airstack-xyz/lib/logger/testing/mocks"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/mock"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	localMock "github.com/airstack-xyz/service-balance-snapshot/testing/mocks"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestNewBalanceSnapshotService(t *testing.T) {
	logger := new(loggerMock.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()

	cache := new(cacheMock.ICache)
	redisTestServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: redisTestServer.Addr(),
	})
	defer redisTestServer.Close()
	distLock := distributedlock.New(client, logger)
	tokenRepo := new(localMock.ITokensRepository)
	balanceSnapshotRepo := new(localMock.IBalanceSnapshotRepository)

	balanceSnapshotService := NewBalanceSnapshotService(logger, cache, tokenRepo, balanceSnapshotRepo, distLock)

	assert.NotNil(t, balanceSnapshotService)
	assert.Equal(t, cache, balanceSnapshotService.cache)
	assert.Equal(t, logger, balanceSnapshotService.logger)
	assert.Equal(t, tokenRepo, balanceSnapshotService.tokenRepo)
	assert.Equal(t, balanceSnapshotRepo, balanceSnapshotService.balanceSnapshotRepo)
	assert.Equal(t, distLock, balanceSnapshotService.distributedlock)
}

func TestGetTokenDataFromTransferEvent(t *testing.T) {
	transfer := getSampleSingleTransfer()
	trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
	transferModel, err := GetTransferFromTransferData(trasferMessage)

	t.Setenv(constants.CHAINID, "1")
	t.Setenv("RPC_ENDPOINT_1", "http://176.9.59.109:8545")

	logger := new(loggerMock.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()
	logger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	cache := new(cacheMock.ICache)
	cache.On("GetObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cache.On("SetObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	redisTestServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: redisTestServer.Addr(),
	})
	defer redisTestServer.Close()
	distLock := distributedlock.New(client, logger)
	tokenRepo := new(localMock.ITokensRepository)
	balanceSnapshotRepo := new(localMock.IBalanceSnapshotRepository)

	balanceSnapshotService := NewBalanceSnapshotService(logger, cache, tokenRepo, balanceSnapshotRepo, distLock)

	token, _, err := balanceSnapshotService.GetTokenDataFromTransferEvent(context.Background(), transferModel)
	assert.Nil(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, transfer.TokenAddress, token.Address)
}
