package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/berikulyBeket/todo-plus/config"
	"github.com/berikulyBeket/todo-plus/internal/consumer"
	handler "github.com/berikulyBeket/todo-plus/internal/controller/http/v1"
	"github.com/berikulyBeket/todo-plus/internal/repository"
	"github.com/berikulyBeket/todo-plus/internal/usecase"
	"github.com/berikulyBeket/todo-plus/pkg/cache"
	"github.com/berikulyBeket/todo-plus/pkg/database"
	"github.com/berikulyBeket/todo-plus/pkg/httpserver"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	messagebroker "github.com/berikulyBeket/todo-plus/pkg/message_broker"
	"github.com/berikulyBeket/todo-plus/pkg/metrics"
	"github.com/berikulyBeket/todo-plus/pkg/search"
)

// Run sets up and starts the application based on the provided configuration
func Run(cfg *config.Config) {
	logger := logger.New(cfg.Log.Level)
	metrics := metrics.New()

	postgresClient, err := initPostgresClient(cfg.Postgres)
	if err != nil {
		logger.Errorf("failed to initialize PostgreSQL client: %v", err)
		return
	}
	defer postgresClient.Close()

	masterRedisClient, replicaRedisClient, err := initRedisClients(cfg.Redis)
	if err != nil {
		logger.Errorf("failed to initialize Redis clients: %v", err)
		return
	}
	defer masterRedisClient.Close()
	defer replicaRedisClient.Close()

	elasticClient, transport, err := initElasticClient(cfg.Elastic)
	if err != nil {
		logger.Errorf("failed to initialize Elastic client: %v", err)
		return
	}
	defer transport.CloseIdleConnections()

	kafkaClient, err := initKafkaClient(cfg.Kafka)
	if err != nil {
		logger.Errorf("failed to initialize Kafka client: %v", err)
		return
	}
	defer kafkaClient.Close()

	db := database.New(postgresClient.DB)
	masterCache := cache.NewRedisCache(masterRedisClient)
	replicaCache := cache.NewRedisCache(replicaRedisClient)
	cache := cache.New(masterCache, replicaCache)
	messageBroker := messagebroker.New(kafkaClient, logger)
	searchService := search.New(elasticClient)
	hasher, tokenMaker := initAuthServices(cfg.AuthSettings)
	appAuth := initAppAuthService(cfg.ApiKeys)

	repos := repository.NewRepository(db, cache, logger)
	usecases := usecase.NewUseCase(repos, searchService, messageBroker.Producer, hasher, tokenMaker, logger)
	handlers := handler.NewHandler(usecases, appAuth, logger, metrics)

	consumers := consumer.New(messageBroker.Consumer, usecases, logger)
	if err := consumers.Init(); err != nil {
		logger.Errorf("failed to start consumers: %v", err)
	}

	router := initRouter(cfg, logger, metrics)

	httpServer := httpserver.New(handlers.RegisterRoutes(router), httpserver.HostPort(cfg.HTTP.Host, cfg.HTTP.Port))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Info("received system signal: " + s.String())
	case err := <-httpServer.Notify():
		logger.Error(fmt.Errorf("error from HTTP server notification: %w", err))
	}

	err = httpServer.Shutdown()
	if err != nil {
		logger.Error(fmt.Errorf("error during HTTP server shutdown: %w", err))
	}
}
