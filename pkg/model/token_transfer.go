package model

import "time"

type TokenTransfer struct {
	ID                          string    `json:"id,omitempty" bson:"_id,omitempty"`
	ChainId                     string    `json:"chainId,omitempty" bson:"chainId,omitempty"`
	Blockchain                  string    `json:"blockchain,omitempty" bson:"blockchain,omitempty"`
	From                        string    `json:"from,omitempty" bson:"from,omitempty"`
	To                          string    `json:"to,omitempty" bson:"to,omitempty"`
	Type                        string    `json:"type,omitempty" bson:"type,omitempty"`
	TokenAddress                string    `json:"tokenAddress,omitempty" bson:"tokenAddress,omitempty"`
	Operator                    string    `json:"operator,omitempty" bson:"operator,omitempty"`
	FormattedAmount             *float64  `json:"formattedAmount,omitempty" bson:"formattedAmount,omitempty"`
	Amount                      string    `json:"amount,omitempty" bson:"amount,omitempty"`
	TokenId                     *string   `json:"tokenId,omitempty" bson:"tokenId,omitempty"`
	Amounts                     []string  `json:"amounts,omitempty" bson:"amounts,omitempty"`
	TokenIds                    []string  `json:"tokenIds,omitempty" bson:"tokenIds,omitempty"`
	TokenType                   string    `json:"tokenType,omitempty" bson:"tokenType,omitempty"`
	TransactionHash             string    `json:"transactionHash,omitempty" bson:"transactionHash,omitempty"`
	FromAddressBalance          *string   `json:"fromAddressBalance,omitempty" bson:"fromAddressBalance,omitempty"`
	FormattedFromAddressBalance *float64  `json:"formattedFromAddressBalance,omitempty" bson:"formattedFromAddressBalance,omitempty"`
	ToAddressBalance            *string   `json:"toAddressBalance,omitempty" bson:"toAddressBalance,omitempty"`
	FormattedToAddressBalance   *float64  `json:"formattedToAddressBalance,omitempty" bson:"formattedToAddressBalance,omitempty"`
	Source                      string    `json:"source,omitempty" bson:"source,omitempty"`
	LogIndex                    int64     `json:"logIndex,omitempty" bson:"logIndex,omitempty"`
	CallIndex                   int64     `json:"callIndex,omitempty" bson:"callIndex,omitempty"`
	CallDepth                   int64     `json:"callDepth,omitempty" bson:"callDepth,omitempty"`
	BlockTimestamp              time.Time `json:"blockTimestamp,omitempty" bson:"blockTimestamp,omitempty"`
	BlockNumber                 int64     `json:"blockNumber,omitempty" bson:"blockNumber,omitempty"`
	CreatedAt                   time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt                   time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
