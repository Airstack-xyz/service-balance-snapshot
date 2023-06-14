package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	helper "github.com/airstack-xyz/database-library/pkg/utils"
	kafkaConstants "github.com/airstack-xyz/kafka/pkg/common/constants"
	kafkaDto "github.com/airstack-xyz/kafka/pkg/common/dto"
	"github.com/airstack-xyz/kafka/pkg/common/schema"
	"github.com/airstack-xyz/kafka/pkg/consumer"
	"github.com/airstack-xyz/kafka/pkg/producer"
	cacheLib "github.com/airstack-xyz/lib/cache"
	distributedlock "github.com/airstack-xyz/lib/distributed-lock"
	"github.com/airstack-xyz/lib/logger"
	"github.com/airstack-xyz/lib/rpc"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/dto"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/model"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/repository"
	repo "github.com/airstack-xyz/service-balance-snapshot/pkg/repository"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/utils"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IBalanceSnapshotService interface {
	SetKafkaReader(reader *consumer.KafkaReader)
	SetKafkaWriter(writer *producer.KafkaWriter)
	ProcessKafkaEventTokenTransfer(ctx context.Context, evt kafkaDto.Message, ch chan consumer.ResponseChan)
	GetTokenDataFromTransferEvent(ctx context.Context, transferTokenData *model.TokenTransfer) (*model.Token, bool, error)
}

type TokenCacheKey struct {
	Address    string `json:"address"`
	Blockchain string `json:"blockchain"`
}

type BalanceSnapshotService struct {
	tokenRepo              repo.ITokensRepository
	balanceSnapshotRepo    repo.IBalanceSnapshotRepository
	distributedlock        *distributedlock.DistributedLock
	logger                 logger.ILogger
	reader                 consumer.IKafkaReader
	writer                 producer.IKafkaWriter
	cache                  cacheLib.ICache
	backfillStartBlockNoAt *uint64
	backfillEndBlockNoAt   *uint64
}

func NewBalanceSnapshotService(logger logger.ILogger, cache cacheLib.ICache, tokenRepo repository.ITokensRepository, balanceSnapshotRepo repository.IBalanceSnapshotRepository,
	distributedLock *distributedlock.DistributedLock) *BalanceSnapshotService {
	return &BalanceSnapshotService{
		logger:              logger,
		cache:               cache,
		tokenRepo:           tokenRepo,
		balanceSnapshotRepo: balanceSnapshotRepo,
		distributedlock:     distributedLock,
	}
}

func (s *BalanceSnapshotService) SetKafkaReader(reader consumer.IKafkaReader) {
	s.reader = reader
}

func (s *BalanceSnapshotService) SetKafkaWriter(writer producer.IKafkaWriter) {
	s.writer = writer
}

func (s *BalanceSnapshotService) SetBackfillProcessingBlockRange(startBlock uint64, endBlock uint64) {
	s.backfillStartBlockNoAt = &startBlock
	s.backfillEndBlockNoAt = &endBlock
}

func (s *BalanceSnapshotService) ProcessKafkaEventTokenTransfer(ctx context.Context, evt kafkaDto.Message, ch chan consumer.ResponseChan) {
	idempotencyKey := evt.Header[constants.IDEMPOTENCYKEY]
	event := &evt
	startTime := time.Now()
	switch event.EventName {
	case kafkaConstants.EVENT_TOKEN_TRANSFERRED, kafkaConstants.EVENT_TOKEN_TRANSFER_1155_SINGLE:
		tokenTransferMessage := event.Value.(*schema.TokenTransferMessage).Value
		transferMessage := CreateTransferMessageFromSingleTransfer(&tokenTransferMessage)
		processTokenTransferError := s.ProcessTokenTransfer(ctx, transferMessage)
		if processTokenTransferError != nil {
			s.logger.Errorf(ctx, "token transfer message processed with error. idempotencyKey: %v . time taken: %s . error: %w", idempotencyKey, time.Since(startTime).String(), processTokenTransferError)
		}
		ch <- consumer.ResponseChan{Err: processTokenTransferError, BatchId: event.BatchId, BatchIndex: event.BatchIndex}
	case kafkaConstants.EVENT_TOKEN_TRANSFER_BATCH:
		tokenTransferMessage := event.Value.(*schema.TokenTransferBatchMessage).Value
		transferMessage := CreateTransferMessageFromBatchTransfer(&tokenTransferMessage)
		processTokenTransferError := s.ProcessTokenTransfer(ctx, transferMessage)
		if processTokenTransferError != nil {
			s.logger.Errorf(ctx, "token transfer message processed with error. idempotencyKey: %v . time taken: %s . error: %w", idempotencyKey, time.Since(startTime).String(), processTokenTransferError)
		}
		ch <- consumer.ResponseChan{Err: processTokenTransferError, BatchId: event.BatchId, BatchIndex: event.BatchIndex}
	default:
		err := errors.New("unsupported event " + event.EventName + "for topic " + event.Topic)
		ch <- consumer.ResponseChan{Err: err, BatchId: event.BatchId, BatchIndex: event.BatchIndex}
	}
}

func (s *BalanceSnapshotService) ProcessTokenTransfer(ctx context.Context, transferTokenData *dto.TokenTransfer) error {
	defer utils.RecordFunctionExecutionTime(ctx, "ProcessTokenTransfer", s.logger)()

	if !shouldProcess(s.backfillEndBlockNoAt, &transferTokenData.BlockNumber) {
		return nil
	}
	transfer, err := GetTransferFromTransferData(transferTokenData)
	if err != nil {
		s.logger.Errorf(ctx, "ProcessTokenTransfer.GetTransferFromTransferData error: %v", err)
		return err
	}

	token, _, err := s.GetTokenDataFromTransferEvent(ctx, transfer)
	if err != nil {
		s.logger.Errorf(ctx, "ProcessTokenTransfer.GetTokenDataFromTransferEvent error: %v", err)
		return err
	}
	if token.Type == constants.TOKEN_TYPE_UNKNOWN {
		return errors.New("unknown token: " + token.Address)
	} else {
		transfer.TokenType = token.Type
		return s.processSnapshot(ctx, transfer, token)
	}

}

func (s *BalanceSnapshotService) GetTokenDataFromTransferEvent(ctx context.Context, transferTokenData *model.TokenTransfer) (*model.Token, bool, error) {
	defer utils.RecordFunctionExecutionTime(ctx, "GetTokenDataFromTransferEvent", s.logger)()
	blockchain, err := GetBlockchainFromChainId(&transferTokenData.ChainId)
	if err != nil {
		return nil, false, err
	}
	cacheKey := EncodeToBase64(TokenCacheKey{Address: transferTokenData.TokenAddress, Blockchain: blockchain})
	var token *model.Token
	err = s.cache.GetObject(ctx, cacheKey, func() (interface{}, int, error) {
		id := blockchain + transferTokenData.TokenAddress
		token, err = s.tokenRepo.GetToken(ctx, id)
		if err != nil {
			return nil, 0, nil
		}
		return token, utils.GetCacheTTL(), nil
	}, &token)

	isNewToken, shouldUpdateCache := false, false

	if token == nil {
		isNewToken = true
		token = &model.Token{
			ID:         transferTokenData.ChainId + transferTokenData.TokenAddress,
			Blockchain: blockchain,
			ChainId:    transferTokenData.ChainId,
			Address:    transferTokenData.TokenAddress,
			Type:       transferTokenData.TokenType,
		}
	}
	if token.Type != constants.TOKEN_TYPE_BASE_TOKEN {
		if isNewToken {
			rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, s.logger)
			defer rpcInstance.Close()

			rpcService := NewRPCService(token, rpcInstance, s.logger)

			tokenDetails, tokenDetailsError := rpcService.GetMissingToken(ctx, transferTokenData, s.logger)
			if tokenDetailsError != nil {
				return nil, false, tokenDetailsError
			}
			if transferTokenData.TokenType == constants.TOKEN_TYPE_UNKNOWN || len(transferTokenData.TokenType) == 0 {
				token.Type = tokenDetails.Type
			}
			token.Name = tokenDetails.Name
			token.Symbol = tokenDetails.Symbol
			token.Decimals = tokenDetails.Decimals
			token.TotalSupply = tokenDetails.TotalSupply
			token.ContractMetaDataURI = tokenDetails.ContractMetaDataURI
			token.BaseURI = tokenDetails.BaseURI
			token.ContractMetaData = tokenDetails.ContractMetaData
		}
	}
	// update cache with latest token data
	if isNewToken || shouldUpdateCache {
		token.LastTransferBlock = uint64(transferTokenData.BlockNumber)
		token.LastTransferHash = transferTokenData.TransactionHash
		token.LastTransferTimestamp = transferTokenData.BlockTimestamp
		if err := s.cache.SetObject(ctx, cacheKey, token, utils.GetCacheTTL()); err != nil {
			s.logger.Errorf(ctx, "Error while setting/updating cache for cachekey: %s", cacheKey)
		}
	}

	return token, isNewToken, nil
}

func (s *BalanceSnapshotService) processSnapshot(ctx context.Context, transfer *model.TokenTransfer, token *model.Token) error {
	defer utils.RecordFunctionExecutionTime(ctx, "processSnapshot", s.logger)()
	if token != nil && token.Type == constants.TOKEN_TRANSFER_TYPE_UNKNOWN {
		return errors.New("unknown token tranfer type")
	}

	//Get respective block snapshot of token
	balances, err := s.GetTokenBalances(ctx, token, transfer)
	if err != nil {
		return err
	}

	if balances == nil {
		return errors.New("processSnapshot: no balances and err")
	}

	for _, balance := range balances {
		mutexName := "mutex-snap-" + fmt.Sprintf("%s-%s-%s-%s", balance.AccountAddress, balance.ContractAddress, balance.TokenType, balance.TokenId)
		var funcErr error
		toBeExecuted := func() {
			var tokenId string
			if token.Type != constants.TOKEN_TYPE_ERC20 {
				tokenId = balance.TokenId
			}

			snapshot, err := s.balanceSnapshotRepo.GetSnapshotByBlockNumber(ctx, transfer.ChainId, balance.AccountAddress, balance.ContractAddress, tokenId, uint(transfer.BlockNumber))
			if err != nil {
				if err == mongo.ErrNoDocuments {
					//Find next block
					nextSnapshot, err := s.balanceSnapshotRepo.FindFirstNearestHighSnapshotRecord(ctx, transfer.ChainId, balance.AccountAddress, tokenId, balance.ContractAddress, uint(transfer.BlockNumber))
					var (
						endBlock     int64
						endTimestamp time.Time
					)
					if err != nil {
						endBlock = math.MaxInt64
						endTimestamp = time.Unix(constants.MAX_TIME_BALANCE_SNAPSHOT, 0)
					} else {
						endBlock = nextSnapshot.StartBlockNumber
						endTimestamp = nextSnapshot.StartBlockTimestamp
					}
					//Insert one
					newSnapshot := model.BalanceSnapshot{
						ID:                  uuid.NewString(),
						Owner:               balance.AccountAddress,
						ChainID:             transfer.ChainId,
						Blockchain:          transfer.Blockchain,
						TokenAddress:        balance.ContractAddress,
						TokenType:           transfer.TokenType,
						StartBlockNumber:    int64(transfer.BlockNumber),
						EndBlockNumber:      endBlock,
						StartBlockTimestamp: transfer.BlockTimestamp,
						EndBlockTimestamp:   endTimestamp,
						Amount:              balance.Balance,
						FormattedAmount:     balance.FormattedBalance,
					}
					if token.Type != constants.TOKEN_TYPE_ERC20 {
						newSnapshot.TokenId = balance.TokenId
					}
					if err := s.balanceSnapshotRepo.CreateSnapshot(ctx, &newSnapshot); err != nil {
						funcErr = err
						return
					}
				} else {
					funcErr = err
					return
				}
			} else {
				exisitingSnapshotUpdate := make(map[string]interface{})
				if snapshot.StartBlockNumber == int64(transfer.BlockNumber) {
					return
				} else {
					exisitingSnapshotUpdate["endBlockNumber"] = transfer.BlockNumber
					exisitingSnapshotUpdate["endBlockTimestamp"] = transfer.BlockTimestamp
				}
				var (
					endBlock     int64
					endTimestamp time.Time
				)

				endBlock = snapshot.EndBlockNumber
				endTimestamp = snapshot.EndBlockTimestamp

				newSnapshot := model.BalanceSnapshot{
					ID:                  uuid.NewString(),
					Owner:               balance.AccountAddress,
					ChainID:             transfer.ChainId,
					Blockchain:          transfer.Blockchain,
					TokenAddress:        balance.ContractAddress,
					TokenType:           transfer.TokenType,
					StartBlockNumber:    int64(transfer.BlockNumber),
					EndBlockNumber:      endBlock,
					StartBlockTimestamp: transfer.BlockTimestamp,
					EndBlockTimestamp:   endTimestamp,
					Amount:              balance.Balance,
					FormattedAmount:     balance.FormattedBalance,
					UpdatedAt:           time.Now().UTC(),
					CreatedAt:           time.Now().UTC(),
				}

				if token.Type != constants.TOKEN_TYPE_ERC20 {
					newSnapshot.TokenId = balance.TokenId
				}

				var writeModels []mongo.WriteModel
				exisitingSnapshotUpdate["updatedAt"] = time.Now()
				updateFieldsBson := helper.ConvertStructToBsonM(exisitingSnapshotUpdate)
				writeModels = append(writeModels, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": snapshot.ID}).SetUpdate(bson.M{"$set": updateFieldsBson}))
				writeModels = append(writeModels, mongo.NewInsertOneModel().SetDocument(newSnapshot))
				if err := s.balanceSnapshotRepo.BulkWriteSnapshot(ctx, writeModels); err != nil {
					funcErr = err
					return
				}
			}
		}
		if s.distributedlock == nil {
			return errors.New("distributed lock is nil")
		}
		if err := s.distributedlock.ExecuteFunction(ctx, mutexName, toBeExecuted, 5*time.Second); err != nil {
			return err
		}
		if funcErr != nil {
			return funcErr
		}
	}
	return nil
}

func (s *BalanceSnapshotService) GetTokenBalances(ctx context.Context, token *model.Token, transfer *model.TokenTransfer) ([]*model.TokenBalanceOutput, error) {
	rpcInstance := rpc.NewRPC([]string{os.Getenv(constants.CHAINID)}, nil)
	defer rpcInstance.Close()

	rpcService := NewRPCService(token, rpcInstance, s.logger)

	rpcService.PrepareTokenBalanceRPCCallData(transfer)

	if err := rpcService.rpcInstance.Call(ctx, rpcService.rpcCallData); err != nil {
		s.logger.Errorf(ctx, "Error while making rpc call: %+v\n", err.Error())
	}

	return rpcService.ProcessTokenBalanceRpcData(ctx, transfer, *token)

}

func shouldProcess(endBlock *uint64, processingBlock *uint64) bool {
	return endBlock != nil && *endBlock+1 <= *processingBlock
}
