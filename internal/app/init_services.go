package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/berikulyBeket/todo-plus/config"
	"github.com/berikulyBeket/todo-plus/internal/middleware"
	appauth "github.com/berikulyBeket/todo-plus/pkg/app_auth"

	"github.com/berikulyBeket/todo-plus/pkg/hash"
	"github.com/berikulyBeket/todo-plus/pkg/kafka"
	"github.com/berikulyBeket/todo-plus/pkg/logger"
	"github.com/berikulyBeket/todo-plus/pkg/metrics"
	"github.com/berikulyBeket/todo-plus/pkg/postgres"
	"github.com/berikulyBeket/todo-plus/pkg/token"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// initPostgresClient initializes a PostgreSQL client
func initPostgresClient(config config.Postgres) (*postgres.Postgres, error) {
	postgresClient, err := postgres.New(config.URL, postgres.MaxPoolSize(config.PoolMax))
	if err != nil {
		return nil, err
	}

	return postgresClient, nil
}

// initRedisClients initializes both master and replica Redis clients using sentinel settings
func initRedisClients(config config.Redis) (*redis.Client, *redis.Client, error) {
	masterRedisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       config.MasterName,
		SentinelAddrs:    strings.Split(config.SentinelAddrs, ","),
		Password:         config.MasterPassword,
		SentinelPassword: config.MasterPassword,
		ReplicaOnly:      false,
	})

	replicaRedisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       config.MasterName,
		SentinelAddrs:    strings.Split(config.SentinelAddrs, ","),
		SentinelPassword: config.MasterPassword,
		ReplicaOnly:      true,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := masterRedisClient.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to master Redis: %w", err)
	}

	if err := replicaRedisClient.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to replica Redis: %w", err)
	}

	return masterRedisClient, replicaRedisClient, nil
}

// initElasticClient initializes an Elasticsearch client
func initElasticClient(config config.Elastic) (*elasticsearch.Client, *http.Transport, error) {
	esMaxIdleConnsPerHost, _ := strconv.Atoi(config.MaxIdleConnsPerHost)
	esResponseHeaderTimeout, _ := time.ParseDuration(config.ResponseHeaderTimeout)
	esDialTimeout, _ := time.ParseDuration(config.DialTimeout)

	transport := &http.Transport{
		MaxIdleConnsPerHost:   esMaxIdleConnsPerHost,
		ResponseHeaderTimeout: esResponseHeaderTimeout,
		DialContext:           (&net.Dialer{Timeout: esDialTimeout}).DialContext,
	}

	elasticCfg := elasticsearch.Config{
		Addresses: strings.Split(config.Addrs, ","),
		Transport: transport,
	}

	elasticClient, err := elasticsearch.NewClient(elasticCfg)
	if err != nil {
		return nil, nil, err
	}

	return elasticClient, transport, nil
}

// initKafkaClient initializes a Kafka client with the provided brokers
func initKafkaClient(config config.Kafka) (*kafka.KafkaClient, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true

	kafkaClient, err := kafka.New(strings.Split(config.Brokers, ","), kafkaConfig)
	if err != nil {
		return nil, err
	}

	return kafkaClient, err
}

// initAppAuthService initializes the application authentication service with the provided API keys
func initAppAuthService(config config.ApiKeys) appauth.Interface {
	return appauth.New(config.AppId, config.AppKey, config.PrivateAppId, config.PrivateAppKey)
}

// initAuthServices initializes the hashing and token services
func initAuthServices(config config.AuthSettings) (hash.Hasher, token.TokenMaker) {
	hasher := hash.New(config.Salt)
	tokenMaker := token.New(config.SigningKey, config.TokenTTL)

	return hasher, tokenMaker
}

// initRouter initializes a new Gin router with middleware for logging, CORS, and request metrics tracking
func initRouter(config *config.Config, logger logger.Interface, metrics metrics.Interface) *gin.Engine {
	router := gin.New()
	router.Use(middleware.PanicRecovery(logger))
	router.Use(middleware.CORS(config.CORS.AllowedOrigins))
	router.Use(middleware.TrackRequestMetrics(metrics))

	return router
}
