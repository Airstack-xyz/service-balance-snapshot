package service

import (
	"os"
	"testing"

	loggerMock "github.com/airstack-xyz/lib/logger/testing/mocks"
	"github.com/airstack-xyz/lib/rpc"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
)

func TestPrepareTokenBalanceRPCCallData(t *testing.T) {
	t.Setenv(constants.CHAINID, "1")
	t.Setenv("RPC_ENDPOINT_1", "http://176.9.59.109:8545")

	logger := new(loggerMock.ILogger)
	logger.On("Info", "successfully connected to MongoDB!").Return()
	logger.On("Info", "mongoDB connection is closed!").Return()
	logger.On("Fatal", mock.Anything, mock.Anything, mock.Anything).Return()
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	t.Run("ERC20 Prepare TokenBalance RPC", func(t *testing.T) {
		transfer := getSampleERC20Transfer()
		trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
		transferModel, _ := GetTransferFromTransferData(trasferMessage)

		token := getSampleERC20Token()

		rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, logger)
		defer rpcInstance.Close()

		rpcService := NewRPCService(&token, rpcInstance, logger)

		rpcService.PrepareTokenBalanceRPCCallData(transferModel)

		assert.Len(t, rpcService.rpcBookKeeping.balanceSnapshot, 2)
		assert.Len(t, rpcService.rpcCallData, 2)
	})

	t.Run("ERC1155 Prepare TokenBalance RPC", func(t *testing.T) {
		transfer := getSampleERC1155BatchTransfer()
		trasferMessage := CreateTransferMessageFromBatchTransfer(&transfer)
		transferModel, _ := GetTransferFromTransferData(trasferMessage)

		token := getSampleERC1155Token()

		rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, logger)
		defer rpcInstance.Close()

		rpcService := NewRPCService(&token, rpcInstance, logger)

		rpcService.PrepareTokenBalanceRPCCallData(transferModel)

		assert.Len(t, rpcService.rpcBookKeeping.balanceSnapshot, 2)
		assert.Len(t, rpcService.rpcCallData, 6)
	})

}
