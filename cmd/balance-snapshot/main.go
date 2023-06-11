package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/airstack-xyz/database-library/pkg/database"
	"github.com/airstack-xyz/kafka/pkg/consumer"
	"github.com/airstack-xyz/kafka/pkg/producer"
	"github.com/airstack-xyz/lib/cache"
	constants_library "github.com/airstack-xyz/lib/constants"
	distributedlock "github.com/airstack-xyz/lib/distributed-lock"
	"github.com/airstack-xyz/lib/logger"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/metrics"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/repository"
	"github.com/airstack-xyz/service-balance-snapshot/pkg/service"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	kafkaMetrics "github.com/airstack-xyz/kafka/pkg/metrics"

	repo "github.com/airstack-xyz/service-balance-snapshot/pkg/repository"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var (
		serverPort  = flag.String("port", "8005", "specify server port")
		environment = flag.String("env", "dev", "specify environment name")
	)

	flag.Parse()
	os.Setenv(constants.ENVIRONMENT, *environment)
	os.Setenv(constants.SERVER_PORT, *serverPort)

	chainId := os.Getenv(constants.CHAINID)
	if len(chainId) == 0 {
		log.Panic("CHAINID env variable not set")
	}

	// instantiate logger
	logger, err := logger.NewLogger()
	if err != nil {
		log.Panic("error while creating logger: ", err)
	}
	defer logger.DisconnectLogger()

	db := database.New(logger)
	if err := db.Connect(ctx, os.Getenv(constants.MONGODB_URI)); err != nil {
		log.Panic("unable to connect to mongodb!")
		log.Panic(err)
	}
	defer func() {
		_ = db.Disconnect(ctx)
	}()

	// init cache
	cache, err := cache.NewCache(logger)
	if err != nil {
		log.Panic("error initializing cache:", err)
	}
	defer cache.Close()

	tokensRepo := repository.NewTokensRepository(db.DB, logger)
	balanceSnapshotRepo := repo.NewBalanceSnapshotRepository(db.DB, logger)
	ds := distributedlock.New(cache.GetClient(), logger)

	balanceSnapshotService := service.NewBalanceSnapshotService(logger, cache, tokensRepo, balanceSnapshotRepo, ds)

	// create producer instance for DLQ
	kafkaWriter, err := producer.Connect(logger)
	if err != nil {
		// handle exception
		log.Panic("error while connecting to broker: ", err)
	}
	defer func() {
		_ = kafkaWriter.Disconnect()
	}()
	balanceSnapshotService.SetKafkaWriter(kafkaWriter)

	// prepare the topic specific configuration
	blockchain := constants_library.ChainIdToBlockchainMap[chainId]
	startOffset, err := strconv.ParseInt(os.Getenv(constants.CONSUMERGROUP_START_OFFSET), 10, 64)
	if err != nil {
		log.Panic("CONSUMERGROUP_START_OFFSET env variable is not set/ not a valid number")
	}
	consumerGroupId := os.Getenv("CONSUMER_GROUPID")
	if len(consumerGroupId) == 0 {
		log.Panic("CONSUMER_GROUPID env variable is not set")
	}
	if chainId != constants.CHAIN_ID_ETHEREUM {
		consumerGroupId = consumerGroupId + "-" + blockchain
	}
	//tokenTopicName := service.GetTopicName(kafkaConstants.TOPIC_TOKEN)
	consumerConfigMap := make(map[string]*consumer.Consumer)

	// transferTopicName := utils.GetTopicName(kafkaConstants.TOPIC_TRANSFER)
	transferTopicName := "test-topic"
	consumerConfigMap[transferTopicName] = &consumer.Consumer{
		GroupID:     consumerGroupId,
		HandlerFunc: balanceSnapshotService.ProcessKafkaEventTokenTransfer,
		StartOffset: startOffset,
	}

	// transferDlqTopicName := os.Getenv("DLQ_TOPIC_NAME")

	// if len(transferDlqTopicName) != 0 {
	// 	consumerConfigMap[transferDlqTopicName] = &consumer.Consumer{
	// 		GroupID:     consumerGroupId,
	// 		HandlerFunc: balanceSnapshotService.ProcessDlqKafkaEventTokenTransfer,
	// 		StartOffset: startOffset,
	// 	}
	// }

	kafkaReader, err := consumer.Subscribe(consumerConfigMap, kafkaWriter.Writer, logger)

	if err != nil {
		log.Panic("unable to create kafka reader instance")
	}
	defer func() {
		_ = kafkaReader.Disconnect()
	}()
	balanceSnapshotService.SetKafkaReader(kafkaReader)

	backfillTillStr := os.Getenv(constants.BACKFILL_TILL_BLOCK_NUMBER)

	if len(backfillTillStr) > 0 {
		if backfillTill, toErr := strconv.ParseUint(backfillTillStr, 10, 64); toErr == nil {
			balanceSnapshotService.SetBackfillProcessingBlockRange(0, backfillTill)
		}
	}

	// below goroutine will take care of polling the kafka message events
	kafkaReader.ConsumeMessages(ctx)

	// registering metrices
	registry := kafkaMetrics.RegisterKafkaMetrics()
	prometheusMiddleWare := func(handler http.Handler) gin.HandlerFunc {
		return func(c *gin.Context) {
			kafkaMetrics.Writer.SetWriterMetrics(kafkaWriter.Writer.Stats())
			for _, consumer := range kafkaReader.ConsumerMap {
				kafkaMetrics.Reader.SetReaderStats(consumer.Reader.Stats())
				kafkaMetrics.Writer.SetWriterMetrics(consumer.Writer.Stats())
			}
			handler.ServeHTTP(c.Writer, c.Request)
		}
	}
	// server init
	localGin := gin.Default()
	pprof.Register(localGin)
	localGin.GET("/health", HealthCheck)
	localGin.GET("/metrics", prometheusMiddleWare(*metrics.GetMetricsHandler(registry)))
	srv := &http.Server{
		Addr:    ":" + *serverPort,
		Handler: localGin,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.

	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}

func HealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "ok")
}
