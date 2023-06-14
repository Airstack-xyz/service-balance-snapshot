package model

import (
	"time"
)

type BalanceSnapshot struct {
	ID                  string    `json:"id" bson:"_id"`
	Owner               string    `json:"owner" bson:"owner"`
	Blockchain          string    `json:"blockchain" bson:"blockchain"`
	ChainID             string    `json:"chainId" bson:"chainId"`
	TokenAddress        string    `json:"tokenAddress" bson:"tokenAddress"`
	TokenId             string    `json:"tokenId,omitempty" bson:"tokenId,omitempty"`
	TokenType           string    `json:"tokenType" bson:"tokenType"`
	StartBlockNumber    int64     `json:"startBlockNumber" bson:"startBlockNumber"`
	EndBlockNumber      int64     `json:"endBlockNumber" bson:"endBlockNumber"`
	StartBlockTimestamp time.Time `json:"startBlockTimestamp" bson:"startBlockTimestamp"`
	EndBlockTimestamp   time.Time `json:"endBlockTimestamp" bson:"endBlockTimestamp"`
	Amount              string    `json:"amount" bson:"amount"`
	FormattedAmount     *float64  `json:"formattedAmount,omitempty" bson:"formattedAmount,omitempty"`
	CreatedAt           time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type TokenBalanceOutput struct {
	Identifier       string
	TokenType        string
	ContractAddress  string
	AccountAddress   string
	TokenId          string
	Balance          string
	FormattedBalance *float64
}
