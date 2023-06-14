package service

import (
	"context"
	"os"
	"testing"

	loggerMock "github.com/airstack-xyz/lib/logger/testing/mocks"
	"github.com/airstack-xyz/lib/rpc"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
)

func TestNewRPCService(t *testing.T) {
	t.Setenv(constants.CHAINID, "1")
	t.Setenv("RPC_ENDPOINT_1", "http://176.9.59.109:8545")

	logger := new(loggerMock.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()
	logger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, logger)
	defer rpcInstance.Close()

	token := getSampleERC20Token()
	rpcService := NewRPCService(&token, rpcInstance, logger)

	assert.NotNil(t, rpcService)
}

func TestProcessTokenBalanceRpcData(t *testing.T) {

	t.Setenv(constants.CHAINID, "1")
	t.Setenv("RPC_ENDPOINT_1", "http://176.9.59.109:8545")

	logger := new(loggerMock.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()
	logger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, logger)
	defer rpcInstance.Close()

	transfer := getSampleERC20Transfer()
	trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
	transferModel, _ := GetTransferFromTransferData(trasferMessage)

	token := getSampleERC20Token()

	rpcService := NewRPCService(&token, rpcInstance, logger)

	rpcService.PrepareTokenBalanceRPCCallData(transferModel)

	rpcService.rpcInstance.Call(context.Background(), rpcService.rpcCallData)

	balanceOutput, err := rpcService.ProcessTokenBalanceRpcData(context.Background(), transferModel, token)

	assert.Nil(t, err)
	assert.NotNil(t, balanceOutput)
	assert.Len(t, balanceOutput, 2)

}

func TestGetMissingToken(t *testing.T) {

	t.Setenv(constants.CHAINID, "1")
	t.Setenv("RPC_ENDPOINT_1", "http://176.9.59.109:8545")

	logger := new(loggerMock.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()
	logger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, logger)
	defer rpcInstance.Close()

	transfer := getSampleERC20Transfer()
	trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
	transferModel, _ := GetTransferFromTransferData(trasferMessage)

	rpcService := NewRPCService(nil, rpcInstance, logger)

	token, err := rpcService.GetMissingToken(context.Background(), transferModel, logger)
	assert.Nil(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, transferModel.TokenAddress, token.Address)
	assert.Equal(t, "ERC20", token.Type)

}
