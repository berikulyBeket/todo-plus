package kafka

import (
	"github.com/IBM/sarama"
)

// KafkaClient wraps a Kafka producer and consumer for message publishing and consuming
type KafkaClient struct {
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
}

// New creates a new KafkaClient with the given brokers
func New(brokers []string, config *sarama.Config) (*KafkaClient, error) {
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		producer.Close()
		return nil, err
	}

	return &KafkaClient{
		Producer: producer,
		Consumer: consumer,
	}, nil
}

// Close closes both the Kafka producer and consumer
func (kc *KafkaClient) Close() error {
	if err := kc.Producer.Close(); err != nil {
		return err
	}
	if err := kc.Consumer.Close(); err != nil {
		return err
	}
	return nil
}
