package service

import (
	"context"
	"strconv"

	"github.com/airstack-xyz/lib/logger"
	"github.com/airstack-xyz/lib/rpc"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
)

type RPCStatus struct {
	id      string
	rpcData *rpc.RPCCallData
}
type RPCBookKeeping struct {
	latestBlockNumber    uint64
	token                map[string]*RPCStatus
	balance              map[string]*RPCStatus
	balanceSnapshot      map[string]*RPCStatus
	owner                map[string]*RPCStatus
	ownerBalanceSnapshot map[string]*RPCStatus
}

type IRPCService interface{}
type RPCService struct {
	Token          *model.Token
	rpcBookKeeping *RPCBookKeeping
	logger         logger.ILogger
	rpcInstance    *rpc.RPC
	rpcCallData    []*rpc.RPCCallData
}

func NewRPCService(
	token *model.Token,
	rpcInstance *rpc.RPC,
	logger logger.ILogger,
) *RPCService {
	return &RPCService{
		Token:          token,
		rpcBookKeeping: &RPCBookKeeping{token: map[string]*RPCStatus{}, balance: map[string]*RPCStatus{}, balanceSnapshot: map[string]*RPCStatus{}, owner: map[string]*RPCStatus{}, ownerBalanceSnapshot: map[string]*RPCStatus{}},
		logger:         logger,
		rpcInstance:    rpcInstance,
		rpcCallData:    []*rpc.RPCCallData{},
	}
}

func (r *RPCService) ProcessTokenBalanceRpcData(ctx context.Context, transfer *model.TokenTransfer, token model.Token) ([]*model.TokenBalanceOutput, error) {
	from := model.TokenBalanceOutput{
		TokenType:       token.Type,
		ContractAddress: token.Address,
		AccountAddress:  transfer.From,
	}
	to := model.TokenBalanceOutput{
		TokenType:       token.Type,
		ContractAddress: token.Address,
		AccountAddress:  transfer.To,
	}

	if transfer.TokenId != nil {
		from.TokenId, to.TokenId = *transfer.TokenId, *transfer.TokenId
	}

	if transfer.TokenType == constants.TOKEN_TYPE_ERC1155 {
		if len(transfer.TokenIds) > 0 {
			var tokenBalanceOutput []*model.TokenBalanceOutput
			for _, tokenId := range transfer.TokenIds {
				// From balance
				fromBalanceOfCurrentToken := from
				fromBalanceOfCurrentToken.TokenId = tokenId
				fromBalance, _ := r.GetERC1155BlockBalanceOfFromAddress(transfer, tokenId)
				fromBalanceOfCurrentToken.Balance = fromBalance
				formattedFromAddressBalance, _ := strconv.ParseFloat(fromBalance, 64)
				fromBalanceOfCurrentToken.FormattedBalance = &formattedFromAddressBalance
				// To balance
				toBalanceOfCurrentToken := to
				toBalanceOfCurrentToken.TokenId = tokenId
				toBalance, _ := r.GetERC1155BlockBalanceOfToAddress(transfer, tokenId)
				toBalanceOfCurrentToken.Balance = toBalance
				formattedToAddressBalance, _ := strconv.ParseFloat(toBalance, 64)
				toBalanceOfCurrentToken.FormattedBalance = &formattedToAddressBalance

				tokenBalanceOutput = append(tokenBalanceOutput, &fromBalanceOfCurrentToken, &toBalanceOfCurrentToken)
			}
			return tokenBalanceOutput, nil
		} else {
			// From balance
			fromBalance, _ := r.GetERC1155BlockBalanceOfFromAddress(transfer, *transfer.TokenId)
			from.Balance = fromBalance
			formattedFromAddressBalance, _ := strconv.ParseFloat(fromBalance, 64)
			from.FormattedBalance = &formattedFromAddressBalance
			// To balance
			toBalance, _ := r.GetERC1155BlockBalanceOfToAddress(transfer, *transfer.TokenId)
			to.Balance = toBalance
			formattedToAddressBalance, _ := strconv.ParseFloat(toBalance, 64)
			to.FormattedBalance = &formattedToAddressBalance
		}
	} else if transfer.TokenType == constants.TOKEN_TYPE_ERC721 {
		// From balance
		fromBalance, _ := r.GetERC721BlockBalanceOfFromAddress(transfer, *transfer.TokenId)
		from.Balance = fromBalance
		formattedFromAddressBalance, _ := strconv.ParseFloat(fromBalance, 64)
		from.FormattedBalance = &formattedFromAddressBalance
		// To balance
		toBalance, _ := r.GetERC721BlockBalanceOfToAddress(transfer, *transfer.TokenId)
		to.Balance = toBalance
		formattedToAddressBalance, _ := strconv.ParseFloat(toBalance, 64)
		to.FormattedBalance = &formattedToAddressBalance
	} else if transfer.TokenType == constants.TOKEN_TYPE_ERC20 {
		// From balance
		fromBalance, _ := r.GetERC20BlockBalanceOfFromAddress(transfer)
		from.Balance = fromBalance
		// To balance
		toBalance, _ := r.GetERC20BlockBalanceOfToAddress(transfer)
		to.Balance = toBalance
		if token.Decimals != nil {
			formattedFromAddressBalance, _ := FormatAmount(fromBalance, *token.Decimals)
			from.FormattedBalance = &formattedFromAddressBalance
			formattedToAddressBalance, _ := FormatAmount(toBalance, *token.Decimals)
			to.FormattedBalance = &formattedToAddressBalance
		}
		// remove tokenId field from transfer if tokenId field has some value
		if transfer.TokenId != nil && len(*transfer.TokenId) > 0 {
			transfer.TokenId = nil
		}
	}

	return []*model.TokenBalanceOutput{&from, &to}, nil

}

func (r *RPCService) GetMissingToken(ctx context.Context, transfer *model.TokenTransfer, logger logger.ILogger) (*model.Token, error) {

	token := model.Token{
		ID:         transfer.ChainId + transfer.TokenAddress,
		Blockchain: transfer.Blockchain,
		ChainId:    transfer.ChainId,
		Address:    transfer.TokenAddress,
		Type:       constants.TOKEN_TYPE_UNKNOWN,
	}

	r.PrepareNewTokenRPCCallData(transfer)

	if err := r.rpcInstance.Call(ctx, r.rpcCallData); err != nil {
		return nil, err
	}

	isERC165Supported, _ := r.SupportERC165(transfer)
	isERC721Supported, _ := r.SupportERC721(transfer)
	isERC1155Supported, _ := r.SupportERC1155(transfer)
	isERCFFFFSupported, _ := r.SupportERCFFFF(transfer)

	// Detect token type
	if isERC165Supported && !isERCFFFFSupported {
		if isERC721Supported {
			token.Type = constants.TOKEN_TYPE_ERC721
			return &token, nil
		} else if isERC1155Supported {
			token.Type = constants.TOKEN_TYPE_ERC1155
			return &token, nil
		}
	}

	// Check the fallback
	tokenURI, _ := r.GetTokenURI(transfer)
	if len(tokenURI) > 0 {
		token.Type = constants.TOKEN_TYPE_ERC721
		return &token, nil
	}

	decimals, _ := r.GetTokenDecimals(transfer)
	if decimals != nil {
		token.Type = constants.TOKEN_TYPE_ERC20
		token.Decimals = decimals
		return &token, nil
	}
	return &token, nil
}
