package service

import (
	"fmt"
	"testing"

	"github.com/airstack-xyz/kafka/pkg/common/schema"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransferMessageFromSingleTransfer(t *testing.T) {
	t.Run("Testing CreateTransferMessageFromSingleTransfer", func(t *testing.T) {
		singleTransfer := getSampleSingleTransfer()
		transferMessage := CreateTransferMessageFromSingleTransfer(&singleTransfer)
		assert.NotNil(t, transferMessage)
		assert.Equal(t, singleTransfer.TransactionHash, transferMessage.TransactionHash)
	})
}

func TestCreateTransferMessageFromBatchTransfer(t *testing.T) {
	t.Run("Testing CreateTransferMessageFromBatchTransfer", func(t *testing.T) {
		batchTransfer := getSampleBatchTransfer()
		transferMessage := CreateTransferMessageFromBatchTransfer(&batchTransfer)
		assert.NotNil(t, transferMessage)
		assert.Equal(t, batchTransfer.TransactionHash, transferMessage.TransactionHash)
	})
}

func TestGetBlockchainFromChainId(t *testing.T) {
	t.Run("Testing GetBlockchainFromChainId", func(t *testing.T) {
		chainId := "1"
		blockchain, err := GetBlockchainFromChainId(&chainId)
		assert.Nil(t, err)
		assert.Equal(t, "ethereum", blockchain)
	})

	t.Run("invalid chainId", func(t *testing.T) {
		chainId := "1020201"
		blockchain, err := GetBlockchainFromChainId(&chainId)
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("unable to map blockchain from chainId %s", chainId))
		assert.Equal(t, "", blockchain)
	})
}

func TestFormatAmount(t *testing.T) {
	t.Run("amount is 0", func(t *testing.T) {
		fmtAmount, err := FormatAmount("0", 10)
		assert.Nil(t, err, nil)
		assert.Equal(t, float64(0), fmtAmount)
	})
	t.Run("amount is empty string", func(t *testing.T) {
		fmtAmount, err := FormatAmount("", 10)
		assert.Nil(t, err, nil)
		assert.Equal(t, float64(0), fmtAmount)
	})
	t.Run("valid case", func(t *testing.T) {
		fmtAmount, err := FormatAmount("1124", 4)
		assert.Nil(t, err, nil)
		assert.Equal(t, float64(0.1124), fmtAmount)
	})
	t.Run("error while converting string to bigInt", func(t *testing.T) {
		_, err := FormatAmount("0.1124", 4)
		assert.NotNil(t, err, nil)
		assert.EqualError(t, err, "error converting,amount=0.1124 to string")
	})
	t.Run("negative number err", func(t *testing.T) {
		_, err := FormatAmount("-1124", 4)
		assert.NotNil(t, err, nil)
		assert.EqualError(t, err, "error converting negative number,amount=-1124")
	})
}

func TestMaxInt(t *testing.T) {
	t.Run("maxInt(a,b) a is greater", func(t *testing.T) {
		a := MaxInt(11, 2)
		assert.Equal(t, 11, a)
	})

	t.Run("maxInt(a,b) b is greater", func(t *testing.T) {
		b := MaxInt(1, 2)
		assert.Equal(t, 2, b)
	})
}

func TestGetTransferType(t *testing.T) {

	t.Run("Normal Transfer", func(t *testing.T) {
		transfer := getSampleSingleTransfer()
		trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
		transferType := GetTransferType(trasferMessage)
		assert.Equal(t, "TRANSFER", transferType)
	})

	t.Run("Mint Transfer", func(t *testing.T) {
		transfer := getSampleSingleTransfer()
		//Making from as ZERO address to test mint transfer
		transfer.From = constants.ZERO_ADDRESS
		trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
		transferType := GetTransferType(trasferMessage)
		assert.Equal(t, "MINT", transferType)
	})

	t.Run("Burn Transfer", func(t *testing.T) {
		transfer := getSampleSingleTransfer()
		//Making to as ZERO address to test Burn transfer
		transfer.To = constants.ZERO_ADDRESS
		trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
		transferType := GetTransferType(trasferMessage)
		assert.Equal(t, "BURN", transferType)
	})

}

func TestGetTransferFromTransferData(t *testing.T) {
	t.Run("get transfer model from transfer message", func(t *testing.T) {
		transfer := getSampleSingleTransfer()
		trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
		transferModel, err := GetTransferFromTransferData(trasferMessage)
		assert.Nil(t, err)
		assert.NotNil(t, transferModel, "transfer model shouldn't be nil")
		assert.Equal(t, "ethereum", transferModel.Blockchain)
		assert.Equal(t, "TRANSFER", transferModel.Type)
	})

	t.Run("wrong chain ID", func(t *testing.T) {
		transfer := getSampleSingleTransfer()
		transfer.ChainId = "100101"
		trasferMessage := CreateTransferMessageFromSingleTransfer(&transfer)
		transferModel, err := GetTransferFromTransferData(trasferMessage)
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("unable to map blockchain from chainId %s", transfer.ChainId))
		assert.Nil(t, transferModel, "transfer model should be nil")
	})
}

func getSampleSingleTransfer() schema.TokenTransfer {
	return schema.TokenTransfer{
		TransactionHash: "0x1459c136ca47579c9201c711989d5bd1346b62ece2e35169a8fa6197cb9af1ff",
		LogIndex:        12,
		CallIndex:       0,
		CallDepth:       0,
		Source:          "LOG",
		ChainId:         "1",
		Operator:        "0x3675a7c40d78cff58492bbc6f72fb829aa8577a2",
		TokenAddress:    "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		TokenId:         "250000000000000000",
		From:            "0xef1c6e67703c7bd7107eed8303fbe6ec2554bf6b",
		To:              "0xea639dfb59d652ab056a2194ff3d9d7ad9744d07",
		Amount:          "250000000000000000",
		TokenType:       "UNKNOWN",
		BlockNumber:     17399294,
		BlockTimestamp:  1685784803,
	}
}

func getSampleBatchTransfer() schema.TokenTransferBatch {
	return schema.TokenTransferBatch{
		TransactionHash: "0x669fda6a3b14c006c65591ee9600d05c2dea139589dc7cf489e1eab083a4e7c5",
		LogIndex:        141,
		Source:          "LOG",
		ChainId:         "1",
		Operator:        "0x9f452b7cc24e6e6fa690fe77cf5dd2ba3dbf1ed9",
		TokenAddress:    "0xc36cf0cfcb5d905b8b513860db0cfe63f6cf9f5c",
		TokenIds: []string{
			"247385280751522262937873339602895489728512",
			"248406127852285078328263463425190794362880",
			"436922559126484987086972995942390383509504",
		},
		From: "0xef10f49704afd226d6af7cfafb9bc7f2f4fc5762",
		To:   "0x9f452b7cc24e6e6fa690fe77cf5dd2ba3dbf1ed9",
		Amounts: []string{
			"1",
			"1",
			"1",
		},
		TokenType:      "ERC1155",
		BlockNumber:    17461068,
		BlockTimestamp: 1686537529,
	}
}
