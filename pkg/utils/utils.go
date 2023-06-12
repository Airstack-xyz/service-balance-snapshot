package utils

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	constants_library "github.com/airstack-xyz/lib/constants"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"

	"github.com/airstack-xyz/lib/logger"
)

func RecordFunctionExecutionTime(ctx context.Context, name string, logger logger.ILogger) func() {
	start := time.Now()
	return func() {
		logger.Debugf(ctx, "%s execution time: %v", name, time.Since(start))
	}
}

func GetChainId() string {
	chainId := os.Getenv(constants.CHAINID)
	return chainId
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
