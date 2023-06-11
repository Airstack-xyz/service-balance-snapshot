package dto

import (
	"math/big"

	"github.com/airstack-xyz/database-library/pkg/model"
	kafkaDto "github.com/airstack-xyz/kafka/pkg/common/dto"
)

type Trace struct {
	TraceId string `json:"trace_id"`
	SpanId  string `json:"span_id"`
}

type Result struct {
	Value interface{}
	Err   error
}

type Balance struct {
	Key            string   `json:"key"`
	ChainId        string   `json:"chain_id"`
	TokenAddress   string   `json:"token_address"`
	TokenId        string   `json:"token_id"`
	AccountAddress string   `json:"account_address"`
	Amount         *big.Int `json:"prev_balance"`
	IsReceiver     bool     `json:"isReceiver"`
	BlockNumber    uint64   `json:"blockNumber"`
	TokenType      string   `json:"tokenType"`
}

type TokenTransfer struct {
	TransactionHash string   `json:"transaction_hash"`
	LogIndex        uint32   `json:"log_index"`
	CallIndex       uint32   `json:"call_index"`
	CallDepth       uint32   `json:"call_depth"`
	Source          string   `json:"source"`
	ChainId         string   `json:"chain_id"`
	Operator        string   `json:"operator"`
	TokenAddress    string   `json:"token_address"`
	TokenId         string   `json:"token_id"`
	TokenIds        []string `json:"token_ids"`
	From            string   `json:"from"`
	To              string   `json:"to"`
	Amount          string   `json:"amount"`
	Amounts         []string `json:"amounts"`
	TokenType       string   `json:"token_type"`
	BlockNumber     uint64   `json:"block_number"`
	BlockTimestamp  uint64   `json:"block_timestamp"`
	IdempotencyKey  string   `json:"idempotency_key"`
}

type ChannelResult struct {
	Identifier string
	Value      interface{}
	Error      error
}

type Channel struct {
	Result chan *ChannelResult
}

type TransferEventDBOperations struct {
	TokenTransfer   map[string]string
	TokenTransferV1 map[string]string
	Token           map[string]string
	TokenBalance    map[string]string
	TokenNFT        map[string]string
}

type TransferEventModels struct {
	TokenTransfer              map[string]*model.TokenTransfer
	TokenTransferV1            map[string]*model.TokenTransfer
	Token                      map[string]*model.Token
	TokenBalance               map[string]*model.TokenBalance
	TokenNFT                   map[string]*model.TokenNft
	ContractMetadataCandidates []string
	ContractMetadata           []kafkaDto.Message
	NFTMetadata                []kafkaDto.Message
	NFTMetadataCandidates      []string
}

type TokenTransferState struct {
	Message   *TokenTransfer
	Operation string
}

type TokenState struct {
	Message   *model.Token
	Operation string
}

type BulkWriteErrors struct {
	TokenTransfer error
	Token         error
	TokenNft      error
	TokenBalance  error
}
