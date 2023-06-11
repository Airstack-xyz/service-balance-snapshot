package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/airstack-xyz/kafka/pkg/common/schema"
	packageConstants "github.com/airstack-xyz/lib/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/dto"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/utils"
	"golang.org/x/exp/slices"
)

func CreateTransferMessageFromSingleTransfer(transfer *schema.TokenTransfer) *dto.TokenTransfer {
	return &dto.TokenTransfer{
		TransactionHash: transfer.TransactionHash,
		LogIndex:        transfer.LogIndex,
		CallIndex:       transfer.CallIndex,
		CallDepth:       transfer.CallDepth,
		Source:          transfer.Source,
		Operator:        transfer.Operator,
		ChainId:         transfer.ChainId,
		TokenAddress:    transfer.TokenAddress,
		TokenId:         transfer.TokenId,
		From:            transfer.From,
		To:              transfer.To,
		Amount:          transfer.Amount,
		TokenType:       transfer.TokenType,
		BlockNumber:     transfer.BlockNumber,
		BlockTimestamp:  transfer.BlockTimestamp,
	}
}

func CreateTransferMessageFromBatchTransfer(transfer *schema.TokenTransferBatch) *dto.TokenTransfer {
	return &dto.TokenTransfer{
		TransactionHash: transfer.TransactionHash,
		LogIndex:        transfer.LogIndex,
		Source:          transfer.Source,
		Operator:        transfer.Operator,
		ChainId:         transfer.ChainId,
		TokenAddress:    transfer.TokenAddress,
		TokenIds:        transfer.TokenIds,
		From:            transfer.From,
		To:              transfer.To,
		Amounts:         transfer.Amounts,
		TokenType:       transfer.TokenType,
		BlockNumber:     transfer.BlockNumber,
		BlockTimestamp:  transfer.BlockTimestamp,
	}
}

func GetBlockchainFromChainId(chainId *string) (string, error) {
	blockchainName := packageConstants.ChainIdToBlockchainMap[*chainId]
	if blockchainName == "" {
		err := errors.New("unable to map blockchain from chainId " + *chainId)
		return "", err
	}
	return blockchainName, nil
}

func FormatAmount(amount string, decimals uint64) (res float64, err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err = fmt.Errorf("recovered from panic format amount, amount: %s decimals: %d", amount, decimals)
		}
	}()
	if len(amount) == 0 {
		return 0, nil
	}
	bigIntAmount, success := new(big.Int).SetString(amount, 10) // converting amount to big.Int
	if !success {
		return 0, fmt.Errorf("error converting,amount=%v to string", amount)
	}
	if bigIntAmount.Sign() == -1 { // if amount < 0
		return 0, fmt.Errorf("error converting negative number,amount=%v", amount)
	}
	if bigIntAmount.Cmp(big.NewInt(0)) == 0 { // if amount is 0 or 00...
		return 0, nil
	}
	bigIntDeno := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil) // 10^decimals
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(bigIntAmount, bigIntDeno, remainder)
	if quotient.Int64() == 0 {
		// 0.{0..}+"amount case"
		amountStr := bigIntAmount.Text(10)
		noOfZeros := MaxInt(0, int(decimals)-len(amountStr))
		numStr := ("0." + strings.Repeat("0", noOfZeros) + amountStr)
		return strconv.ParseFloat(numStr, 64)
	}
	// if remainder is zero
	if remainder.Cmp(big.NewInt(0)) == 0 {
		return strconv.ParseFloat(quotient.Text(10), 64)
	}
	prefixZeros := int(decimals) - len(remainder.Text(10))
	if prefixZeros > 0 {
		numStr := (quotient.Text(10) + "." + strings.Repeat("0", prefixZeros)) + remainder.Text(10)
		return strconv.ParseFloat(numStr, 64)
	}
	numStr := quotient.Text(10) + "." + remainder.Text(10)
	return strconv.ParseFloat(numStr, 64)
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func EncodeToBase64(input interface{}) string {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(inputBytes)
}

func GetTransferType(tokenTransfer *dto.TokenTransfer) string {
	burnAddress := []string{
		"0x000000000000000000000000000000000000dead",
		"0x0000000000000000000000000000000000000000",
		"0x0000000000000000000000000000000000000001",
		"0x0000000000000000000000000000000000000002",
		"0x0000000000000000000000000000000000000003",
		"0x0000000000000000000000000000000000000004",
		"0x0000000000000000000000000000000000000005",
		"0x0000000000000000000000000000000000000006",
		"0x0000000000000000000000000000000000000007",
		"0x0000000000000000000000000000000000000008",
		"0x0000000000000000000000000000000000000009",
		"0x00000000000000000000045261d4ee77acdb3286",
		"0x0123456789012345678901234567890123456789",
		"0x1111111111111111111111111111111111111111",
		"0x1234567890123456789012345678901234567890",
		"0x2222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333",
		"0x4444444444444444444444444444444444444444",
		"0x6666666666666666666666666666666666666666",
		"0x8888888888888888888888888888888888888888",
		"0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"0xdead000000000000000042069420694206942069",
		"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		"0xffffffffffffffffffffffffffffffffffffffff",
		"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	transferType := constants.TOKEN_TRANSFER_TYPE_UNKNOWN
	if slices.Contains(burnAddress, tokenTransfer.From) && !slices.Contains(burnAddress, tokenTransfer.To) {
		transferType = constants.TOKEN_TRANSFER_TYPE_MINT
	} else if slices.Contains(burnAddress, tokenTransfer.To) && !slices.Contains(burnAddress, tokenTransfer.From) {
		transferType = constants.TOKEN_TRANSFER_TYPE_BURN
	} else if !slices.Contains(burnAddress, tokenTransfer.From) && !slices.Contains(burnAddress, tokenTransfer.To) {
		transferType = constants.TOKEN_TRANSFER_TYPE_TRANSFER
	}
	return transferType
}

func GetTransferFromTransferData(transferMessage *dto.TokenTransfer) (*model.TokenTransfer, error) {

	blockchain, err := GetBlockchainFromChainId(&transferMessage.ChainId)
	if err != nil {
		return nil, err
	}

	transferType := utils.GetTransferType(transferMessage)
	transfer := &model.TokenTransfer{
		Blockchain:      blockchain,
		ChainId:         transferMessage.ChainId,
		From:            transferMessage.From,
		To:              transferMessage.To,
		Type:            transferType,
		TokenAddress:    transferMessage.TokenAddress,
		Operator:        transferMessage.Operator,
		Amount:          transferMessage.Amount,
		Amounts:         transferMessage.Amounts,
		TokenId:         &transferMessage.TokenId,
		TokenIds:        transferMessage.TokenIds,
		TokenType:       transferMessage.TokenType,
		TransactionHash: transferMessage.TransactionHash,
		BlockTimestamp:  time.Unix(int64(transferMessage.BlockTimestamp), 0),
		BlockNumber:     int64(transferMessage.BlockNumber),
		LogIndex:        int64(transferMessage.LogIndex),
		Source:          transferMessage.Source,
		CallIndex:       int64(transferMessage.CallIndex),
		CallDepth:       int64(transferMessage.CallDepth),
	}
	return transfer, nil
}
