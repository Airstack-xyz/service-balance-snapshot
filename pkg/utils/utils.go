package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	constants_library "github.com/airstack-xyz/lib/constants"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/dto"
	"golang.org/x/exp/slices"

	"github.com/airstack-xyz/lib/logger"
)

var chainId string

func RecordFunctionExecutionTime(ctx context.Context, name string, logger logger.ILogger) func() {
	start := time.Now()
	return func() {
		logger.Debugf(ctx, "%s execution time: %v", name, time.Since(start))
	}
}

func GenerateHashedID(args ...string) string {
	var finalKey string
	for _, arg := range args {
		finalKey = finalKey + arg
	}
	hash := sha256.Sum256([]byte(finalKey))
	return hex.EncodeToString(hash[:])
}

func GenerateTokenBalanceId(chainId string, tokenAddress string, ownerAddress string, tokenId string) string {
	id := fmt.Sprintf("%s%s%s%s", chainId, tokenAddress, ownerAddress, tokenId)
	if GetChainId() != constants.CHAIN_ID_ETHEREUM {
		id = GenerateHashedID(id)
	}
	return id
}

func GetChainId() string {
	if len(chainId) > 0 {
		return chainId
	}
	chainId = os.Getenv(constants.CHAINID)
	return chainId
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

func GetCacheTTL() int {
	ttl := os.Getenv(constants.CACHE_TTL)
	ttlInt, err := strconv.Atoi(ttl)
	if err != nil {
		return constants.DEFAULT_CACHE_TTL //default
	}
	return ttlInt
}

func Ptr[T any](x T) *T {
	return &x
}

func GetTopicName(topicName string) string {
	topicNameFromEnv := os.Getenv(topicName)
	if topicNameFromEnv != "" {
		return topicNameFromEnv
	}
	chainId := os.Getenv(constants.CHAINID)
	if chainId == constants.CHAIN_ID_ETHEREUM { // returning direct topic without prefix of blockchain for ethereum
		return topicName
	}
	blockchain, err := GetBlockchainFromChainId(&chainId)
	if err != nil {
		log.Panicf("blockchain doesn't exist in constants lib for chainId: %s", chainId)
	}
	return blockchain + "_" + topicName
}

func GetBlockchainFromChainId(chainId *string) (string, error) {
	blockchainName := constants_library.ChainIdToBlockchainMap[*chainId]
	if blockchainName == "" {
		err := errors.New("unable to map blockchain from chainId " + *chainId)
		return "", err
	}
	return blockchainName, nil
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
