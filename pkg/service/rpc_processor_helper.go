package service

import (
	"errors"
	"math/big"
	"strings"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
)

func (r *RPCService) PrepareTokenBalanceRPCCallData(transfer *model.TokenTransfer) {
	tokenId := transfer.TokenId

	if transfer.TokenType == constants.TOKEN_TYPE_ERC1155 {
		if len(transfer.TokenIds) > 0 {
			// This is a batch transfer
			for _, tokenId := range transfer.TokenIds {
				if transfer.To != constants.ZERO_ADDRESS {
					r.PrepareERC1155TokenBalanceRPCCallData(transfer, transfer.To, tokenId)
				}
				if transfer.From != constants.ZERO_ADDRESS {
					r.PrepareERC1155TokenBalanceRPCCallData(transfer, transfer.From, tokenId)
				}
			}
		} else {
			// This is a single erc 1155 transfer
			if transfer.To != constants.ZERO_ADDRESS {
				r.PrepareERC1155TokenBalanceRPCCallData(transfer, transfer.To, *tokenId)
			}

			if transfer.From != constants.ZERO_ADDRESS {
				r.PrepareERC1155TokenBalanceRPCCallData(transfer, transfer.From, *tokenId)
			}
		}
	} else {
		// This can be erc20 or erc721 transfers
		if r.Token == nil {
			r.PrepareERC20TokenBalanceRPCCallData(transfer, transfer.To)
			r.PrepareERC20TokenBalanceRPCCallData(transfer, transfer.From)
			r.PrepareERC721TokenBalanceRPCCallData(transfer, transfer.To, *tokenId)
			r.PrepareERC721TokenBalanceRPCCallData(transfer, transfer.From, *tokenId)
		} else {
			token := r.Token
			if token.Type == constants.TOKEN_TYPE_ERC20 {
				r.PrepareERC20TokenBalanceRPCCallData(transfer, transfer.To)
				r.PrepareERC20TokenBalanceRPCCallData(transfer, transfer.From)
			} else if token.Type == constants.TOKEN_TYPE_ERC721 {
				r.PrepareERC721TokenBalanceRPCCallData(transfer, transfer.To, *tokenId)
				r.PrepareERC721TokenBalanceRPCCallData(transfer, transfer.From, *tokenId)
			}
		}
	}

}

func (r *RPCService) PrepareERC20TokenBalanceRPCCallData(transfer *model.TokenTransfer, ownerAddress string) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	// For balance snapshot
	tokenBalanceSnapshotRPCCallData := r.rpcInstance.GetTokenBalanceCallData(chainId, tokenAddress, ownerAddress, &transfer.BlockNumber)
	tokenBalanceSnapshotRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenBalanceSnapshotRPCCallData)
	r.rpcBookKeeping.balanceSnapshot[tokenBalanceSnapshotRPCCallData.Identifier] = &RPCStatus{
		id:      tokenBalanceSnapshotRPCCallData.Identifier,
		rpcData: tokenBalanceSnapshotRPCCallData,
	}
}

func (r *RPCService) PrepareERC721TokenBalanceRPCCallData(transfer *model.TokenTransfer, ownerAddress string, tokenId string) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	// Balance snapshot
	tokenOwnerSnapshotRPCCallData := r.rpcInstance.GetNft721TokenOwnerCallData(chainId, tokenAddress, tokenId, &transfer.BlockNumber)
	tokenOwnerSnapshotRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenOwnerSnapshotRPCCallData)
	r.rpcBookKeeping.balanceSnapshot[tokenOwnerSnapshotRPCCallData.Identifier] = &RPCStatus{
		id:      tokenOwnerSnapshotRPCCallData.Identifier,
		rpcData: tokenOwnerSnapshotRPCCallData,
	}

}

func (r *RPCService) PrepareERC1155TokenBalanceRPCCallData(transfer *model.TokenTransfer, ownerAddress string, tokenId string) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	// For balance snapshot
	tokenBalanceSnapshotRPCCallData := r.rpcInstance.GetToken1155BalanceCallData(chainId, tokenAddress, ownerAddress, tokenId, &transfer.BlockNumber)
	tokenBalanceSnapshotRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenBalanceSnapshotRPCCallData)
	r.rpcBookKeeping.balanceSnapshot[tokenBalanceSnapshotRPCCallData.Identifier] = &RPCStatus{
		id:      tokenBalanceSnapshotRPCCallData.Identifier,
		rpcData: tokenBalanceSnapshotRPCCallData,
	}
}

func (r *RPCService) GetERC20BalanceOfToAddress(transfer *model.TokenTransfer) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetTokenBalanceCallDataIdentifier(chainId, tokenAddress, transfer.To, nil)
	tokenRpcData := r.rpcBookKeeping.balance[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	balance, ok := tokenRpcData.rpcData.Value.(*big.Int)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return balance.String(), nil
}
func (r *RPCService) GetERC20BalanceOfFromAddress(transfer *model.TokenTransfer) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetTokenBalanceCallDataIdentifier(chainId, tokenAddress, transfer.From, nil)
	tokenRpcData := r.rpcBookKeeping.balance[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	balance, ok := tokenRpcData.rpcData.Value.(*big.Int)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return balance.String(), nil
}
func (r *RPCService) GetERC20BlockBalanceOfToAddress(transfer *model.TokenTransfer) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetTokenBalanceCallDataIdentifier(chainId, tokenAddress, transfer.To, &transfer.BlockNumber)
	tokenRpcData := r.rpcBookKeeping.balanceSnapshot[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	balance, ok := tokenRpcData.rpcData.Value.(*big.Int)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return balance.String(), nil
}
func (r *RPCService) GetERC20BlockBalanceOfFromAddress(transfer *model.TokenTransfer) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetTokenBalanceCallDataIdentifier(chainId, tokenAddress, transfer.From, &transfer.BlockNumber)
	tokenRpcData := r.rpcBookKeeping.balanceSnapshot[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	balance, ok := tokenRpcData.rpcData.Value.(*big.Int)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return balance.String(), nil
}

func (r *RPCService) GetERC1155BlockBalanceOfToAddress(transfer *model.TokenTransfer, tokenId string) (string, error) {
	if transfer.To == constants.ZERO_ADDRESS {
		return "0", nil
	}
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetToken1155BalanceCallDataIdentifier(chainId, tokenAddress, transfer.To, tokenId, &transfer.BlockNumber)
	tokenRpcData := r.rpcBookKeeping.balanceSnapshot[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	balance, ok := tokenRpcData.rpcData.Value.(*big.Int)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return balance.String(), nil
}
func (r *RPCService) GetERC1155BlockBalanceOfFromAddress(transfer *model.TokenTransfer, tokenId string) (string, error) {
	if transfer.From == constants.ZERO_ADDRESS {
		return "0", nil
	}
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetToken1155BalanceCallDataIdentifier(chainId, tokenAddress, transfer.From, tokenId, &transfer.BlockNumber)
	tokenRpcData := r.rpcBookKeeping.balanceSnapshot[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	balance, ok := tokenRpcData.rpcData.Value.(*big.Int)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return balance.String(), nil
}

func (r *RPCService) GetERC721BalanceOfToAddress(transfer *model.TokenTransfer, tokenId string) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetNft721TokenOwnerCallDataIdentifier(chainId, tokenAddress, tokenId, nil)
	tokenRpcData := r.rpcBookKeeping.balance[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	ownerAddress, ok := tokenRpcData.rpcData.Value.(string)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	if strings.EqualFold(ownerAddress, transfer.To) {
		return "1", nil
	}
	return "0", nil
}
func (r *RPCService) GetERC721BalanceOfFromAddress(transfer *model.TokenTransfer, tokenId string) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetNft721TokenOwnerCallDataIdentifier(chainId, tokenAddress, tokenId, nil)
	tokenRpcData := r.rpcBookKeeping.balance[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	ownerAddress, ok := tokenRpcData.rpcData.Value.(string)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	if strings.EqualFold(ownerAddress, transfer.From) {
		return "1", nil
	}
	return "0", nil
}
func (r *RPCService) GetERC721BlockBalanceOfToAddress(transfer *model.TokenTransfer, tokenId string) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetNft721TokenOwnerCallDataIdentifier(chainId, tokenAddress, tokenId, &transfer.BlockNumber)
	tokenRpcData := r.rpcBookKeeping.balanceSnapshot[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	ownerAddress, ok := tokenRpcData.rpcData.Value.(string)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	if strings.EqualFold(ownerAddress, transfer.To) {
		return "1", nil
	}
	return "0", nil
}
func (r *RPCService) GetERC721BlockBalanceOfFromAddress(transfer *model.TokenTransfer, tokenId string) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	id := r.rpcInstance.GetNft721TokenOwnerCallDataIdentifier(chainId, tokenAddress, tokenId, &transfer.BlockNumber)
	tokenRpcData := r.rpcBookKeeping.balanceSnapshot[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	ownerAddress, ok := tokenRpcData.rpcData.Value.(string)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	if strings.EqualFold(ownerAddress, transfer.From) {
		return "1", nil
	}
	return "0", nil
}

func (r *RPCService) GetTokenDecimals(transfer *model.TokenTransfer) (*uint64, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	tokenDecimalsId := r.rpcInstance.GetTokenDecimalsCallDataIdentifier(chainId, tokenAddress, nil)
	tokenRpcData := r.rpcBookKeeping.token[tokenDecimalsId]
	if tokenRpcData != nil && tokenRpcData.rpcData.Value != nil {
		decimalBN, ok := tokenRpcData.rpcData.Value.(uint8)
		if !ok {
			err := errors.New("decimals should be of type uint8")
			return nil, err
		}
		decimals := uint64(decimalBN)
		return &decimals, nil
	}
	return nil, errors.New("no source available")
}

func (r *RPCService) PrepareNewTokenRPCCallData(transfer *model.TokenTransfer) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress

	// We also need to perform the token type detection so call interface calls
	erc165SupportRPCCallData := r.rpcInstance.GetSupportERC165CallData(chainId, tokenAddress, nil)
	erc165SupportRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, erc165SupportRPCCallData)
	r.rpcBookKeeping.token[erc165SupportRPCCallData.Identifier] = &RPCStatus{
		id:      erc165SupportRPCCallData.Identifier,
		rpcData: erc165SupportRPCCallData,
	}

	erc721SupportRPCCallData := r.rpcInstance.GetSupportERC721CallData(chainId, tokenAddress, nil)
	erc721SupportRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, erc721SupportRPCCallData)
	r.rpcBookKeeping.token[erc721SupportRPCCallData.Identifier] = &RPCStatus{
		id:      erc721SupportRPCCallData.Identifier,
		rpcData: erc721SupportRPCCallData,
	}

	erc1155SupportRPCCallData := r.rpcInstance.GetSupportERC1155CallData(chainId, tokenAddress, nil)
	erc1155SupportRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, erc1155SupportRPCCallData)
	r.rpcBookKeeping.token[erc1155SupportRPCCallData.Identifier] = &RPCStatus{
		id:      erc1155SupportRPCCallData.Identifier,
		rpcData: erc1155SupportRPCCallData,
	}

	ercFFFFSupportRPCCallData := r.rpcInstance.GetSupportERCFFFFCallData(chainId, tokenAddress, nil)
	ercFFFFSupportRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, ercFFFFSupportRPCCallData)
	r.rpcBookKeeping.token[ercFFFFSupportRPCCallData.Identifier] = &RPCStatus{
		id:      ercFFFFSupportRPCCallData.Identifier,
		rpcData: ercFFFFSupportRPCCallData,
	}

	// The token does not exists in the DB. So get all the details.
	tokenNameRPCCallData := r.rpcInstance.GetTokenNameCallData(chainId, tokenAddress, nil)
	tokenNameRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenNameRPCCallData)
	r.rpcBookKeeping.token[tokenNameRPCCallData.Identifier] = &RPCStatus{
		id:      tokenNameRPCCallData.Identifier,
		rpcData: tokenNameRPCCallData,
	}

	tokenSymbolRPCCallData := r.rpcInstance.GetTokenSymbolCallData(chainId, tokenAddress, nil)
	tokenSymbolRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenSymbolRPCCallData)
	r.rpcBookKeeping.token[tokenSymbolRPCCallData.Identifier] = &RPCStatus{
		id:      tokenSymbolRPCCallData.Identifier,
		rpcData: tokenSymbolRPCCallData,
	}

	tokenDecimalsRPCCallData := r.rpcInstance.GetTokenDecimalsCallData(chainId, tokenAddress, nil)
	tokenDecimalsRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenDecimalsRPCCallData)
	r.rpcBookKeeping.token[tokenDecimalsRPCCallData.Identifier] = &RPCStatus{
		id:      tokenDecimalsRPCCallData.Identifier,
		rpcData: tokenDecimalsRPCCallData,
	}

	tokenTotalSupplyRPCCallData := r.rpcInstance.GetTokenTotalSupplyCallData(chainId, tokenAddress, nil)
	tokenTotalSupplyRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenTotalSupplyRPCCallData)
	r.rpcBookKeeping.token[tokenTotalSupplyRPCCallData.Identifier] = &RPCStatus{
		id:      tokenTotalSupplyRPCCallData.Identifier,
		rpcData: tokenTotalSupplyRPCCallData,
	}

	tokenContractUriRPCCallData := r.rpcInstance.GetTokenContractURICallData(chainId, tokenAddress, nil)
	tokenContractUriRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenContractUriRPCCallData)
	r.rpcBookKeeping.token[tokenContractUriRPCCallData.Identifier] = &RPCStatus{
		id:      tokenContractUriRPCCallData.Identifier,
		rpcData: tokenContractUriRPCCallData,
	}

	tokenBaseUriRPCCallData := r.rpcInstance.GetNft721BaseURICallData(chainId, tokenAddress, nil)
	tokenBaseUriRPCCallData.Ref = transfer.ID
	r.rpcCallData = append(r.rpcCallData, tokenBaseUriRPCCallData)
	r.rpcBookKeeping.token[tokenBaseUriRPCCallData.Identifier] = &RPCStatus{
		id:      tokenBaseUriRPCCallData.Identifier,
		rpcData: tokenBaseUriRPCCallData,
	}
}

func (r *RPCService) GetTokenURI(transfer *model.TokenTransfer) (string, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	tokenId := transfer.TokenId
	if tokenId == nil || len(*tokenId) == 0 {
		tokenId = &transfer.Amount
	}

	id := r.rpcInstance.GetNft721TokenURICallDataIdentifier(chainId, tokenAddress, *tokenId, nil)
	tokenRpcData := r.rpcBookKeeping.token[id]
	if tokenRpcData == nil || tokenRpcData.rpcData.Value == nil {
		return "", errors.New("no rpc data available")
	}

	tokenURI, ok := tokenRpcData.rpcData.Value.(string)
	if !ok {
		err := errors.New("unable to convert to string")
		return "", err
	}
	return tokenURI, nil
}

func (r *RPCService) SupportERC721(transfer *model.TokenTransfer) (bool, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	id := r.rpcInstance.GetSupportERC721CallDataIdentifier(chainId, tokenAddress, nil)
	tokenRpcData := r.rpcBookKeeping.token[id]
	if tokenRpcData != nil && tokenRpcData.rpcData.Value != nil {
		supports, ok := tokenRpcData.rpcData.Value.(bool)
		if !ok {
			err := errors.New("unable to convert to bool")
			return false, err
		}
		return supports, nil
	}
	return false, nil
}

func (r *RPCService) SupportERC1155(transfer *model.TokenTransfer) (bool, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	id := r.rpcInstance.GetSupportERC1155CallDataIdentifier(chainId, tokenAddress, nil)
	tokenRpcData := r.rpcBookKeeping.token[id]
	if tokenRpcData != nil && tokenRpcData.rpcData.Value != nil {
		supports, ok := tokenRpcData.rpcData.Value.(bool)
		if !ok {
			err := errors.New("unable to convert to bool")
			return false, err
		}
		return supports, nil
	}
	return false, nil
}

func (r *RPCService) SupportERCFFFF(transfer *model.TokenTransfer) (bool, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	id := r.rpcInstance.GetSupportERCFFFFCallDataIdentifier(chainId, tokenAddress, nil)
	tokenRpcData := r.rpcBookKeeping.token[id]
	if tokenRpcData != nil && tokenRpcData.rpcData.Value != nil {
		supports, ok := tokenRpcData.rpcData.Value.(bool)
		if !ok {
			err := errors.New("unable to convert to bool")
			return false, err
		}
		return supports, nil
	}
	return false, nil
}

func (r *RPCService) SupportERC165(transfer *model.TokenTransfer) (bool, error) {
	chainId := transfer.ChainId
	tokenAddress := transfer.TokenAddress
	id := r.rpcInstance.GetSupportERC165CallDataIdentifier(chainId, tokenAddress, nil)
	tokenRpcData := r.rpcBookKeeping.token[id]
	if tokenRpcData != nil && tokenRpcData.rpcData.Value != nil {
		supports, ok := tokenRpcData.rpcData.Value.(bool)
		if !ok {
			err := errors.New("unable to convert to bool")
			return false, err
		}
		return supports, nil
	}
	return false, nil
}
