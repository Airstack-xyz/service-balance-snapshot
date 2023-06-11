package model

import (
	"time"
)

type TokenBalance struct {
	ID                   string    `json:"id,omitempty" bson:"_id,omitempty"`
	Blockchain           string    `json:"blockchain,omitempty" bson:"blockchain,omitempty"`
	Owner                string    `json:"owner,omitempty" bson:"owner,omitempty"`
	ChainId              string    `json:"chainId,omitempty" bson:"chainId,omitempty"`
	TokenAddress         string    `json:"tokenAddress,omitempty" bson:"tokenAddress,omitempty"`
	TokenType            string    `json:"tokenType,omitempty" bson:"tokenType,omitempty"`
	TokenId              string    `json:"tokenId,omitempty" bson:"tokenId,omitempty"`
	FormattedAmount      float64   `json:"formattedAmount,omitempty" bson:"formattedAmount,omitempty"`
	Amount               string    `json:"amount,omitempty" bson:"amount,omitempty"`
	LastUpdatedBlock     uint64    `json:"lastUpdatedBlock,omitempty" bson:"lastUpdatedBlock,omitempty"`
	LastUpdatedTimestamp time.Time `json:"lastUpdatedTimestamp,omitempty" bson:"lastUpdatedTimestamp,omitempty"`
	LastTransactionHash  string    `json:"lastTransactionHash,omitempty" bson:"lastTransactionHash,omitempty"`
	CreatedAt            time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt            time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}




