package model

import (
	"time"
)

type LogoSizes struct {
	Small    string `json:"small,omitempty" bson:"small,omitempty"`
	Medium   string `json:"medium,omitempty" bson:"medium,omitempty"`
	Large    string `json:"large,omitempty" bson:"large,omitempty"`
	Original string `json:"original,omitempty" bson:"original,omitempty"`
}

type Token struct {
	ID                        string                 `json:"id,omitempty" bson:"_id,omitempty"`
	Blockchain                string                 `json:"blockchain,omitempty" bson:"blockchain,omitempty"`
	Address                   string                 `json:"address,omitempty" bson:"address,omitempty"`
	ChainId                   string                 `json:"chainId,omitempty" bson:"chainId,omitempty"`
	Name                      string                 `json:"name,omitempty" bson:"name,omitempty"`
	Symbol                    string                 `json:"symbol,omitempty" bson:"symbol,omitempty"`
	Type                      string                 `json:"type,omitempty" bson:"type,omitempty"`
	TotalSupply               string                 `json:"totalSupply,omitempty" bson:"totalSupply,omitempty"`
	Decimals                  *uint64                `json:"decimals,omitempty" bson:"decimals,omitempty"`
	Logo                      LogoSizes              `json:"logo,omitempty" bson:"logo,omitempty"`
	ContractMetaDataURI       string                 `json:"contractMetaDataURI,omitempty" bson:"contractMetaDataURI,omitempty"`
	ContractMetaData          map[string]interface{} `json:"contractMetaData,omitempty" bson:"contractMetaData,omitempty"`
	BaseURI                   string                 `json:"baseURI,omitempty" bson:"baseURI,omitempty"`
	LastTransferTimestamp     time.Time              `json:"lastTransferTimestamp,omitempty" bson:"lastTransferTimestamp,omitempty"`
	LastTransferBlock         uint64                 `json:"lastTransferBlock,omitempty" bson:"lastTransferBlock,omitempty"`
	LastTransferHash          string                 `json:"lastTransferHash,omitempty" bson:"lastTransferHash,omitempty"`
	CurrentHolderCount        int64                  `json:"currentHolderCount,omitempty" bson:"currentHolderCount,omitempty"`
	TransferCount             uint64                 `json:"transferCount,omitempty" bson:"transferCount,omitempty"`
	DeploymentTransactionHash string                 `json:"deploymentTransactionHash,omitempty" bson:"deploymentTransactionHash,omitempty"`
	DeployedAtBlock           uint64                 `json:"deployedAtBlock,omitempty" bson:"deployedAtBlock,omitempty"`
	Deployer                  string                 `json:"deployer,omitempty" bson:"deployer,omitempty"`
	DeployedAt                time.Time              `json:"deployedAt,omitempty" bson:"deployedAt,omitempty"`
	TokenTraits               map[string]interface{} `json:"tokenTraits,omitempty" bson:"tokenTraits,omitempty"`
	CreatedAt                 time.Time              `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt                 time.Time              `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}
