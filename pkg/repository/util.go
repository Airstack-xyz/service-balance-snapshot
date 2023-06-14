package repository

import (
	"log"
	"strings"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/utils"

	constants_library "github.com/airstack-xyz/lib/constants"
)

func getDB(dbName string) string {
	chainId := utils.GetChainId()
	if chainId == constants.CHAIN_ID_ETHEREUM { // returning direct topic without prefix of blockchain for ethereum
		return dbName
	}
	blockchain := constants_library.ChainIdToBlockchainMap[chainId]
	if blockchain == constants.EMPTY_STRING {
		log.Panicf("blockchain: %s doesn't exist in constants lib", blockchain)
	}
	return strings.ToUpper(blockchain) + "_" + dbName
}
