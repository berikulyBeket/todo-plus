package messagebroker

import (
	"encoding/json"

	"github.com/berikulyBeket/todo-plus/internal/entity"
	"github.com/berikulyBeket/todo-plus/pkg/kafka"
	"github.com/berikulyBeket/todo-plus/pkg/logger"

	"github.com/IBM/sarama"
)

// Constants defining Kafka topic names for various events
const (
	ListCreatedTopic = "list_created"
	ListUpdatedTopic = "list_updated"
	ListDeletedTopic = "list_deleted"
	ItemCreatedTopic = "item_created"
	ItemUpdatedTopic = "item_updated"
	ItemDeletedTopic = "item_deleted"
)

// KafkaBroker wraps both Kafka producer and consumer
type KafkaBroker struct {
	Producer
	Consumer
}

// KafkaProducer implements the Producer interface for publishing events to Kafka
type KafkaProducer struct {
	client *kafka.KafkaClient
	logger logger.Interface
}

// KafkaConsumer implements the Consumer interface for consuming events from Kafka
type KafkaConsumer struct {
	client *kafka.KafkaClient
	logger logger.Interface
}

// New initializes a new KafkaBroker with a Kafka client and logger
func New(client *kafka.KafkaClient, logger logger.Interface) *KafkaBroker {
	return &KafkaBroker{
		Producer: NewKafkaProducer(client, logger),
		Consumer: NewKafkaConsumer(client, logger),
	}
}

// NewKafkaProducer creates a new KafkaProducer instance
func NewKafkaProducer(client *kafka.KafkaClient, logger logger.Interface) Producer {
	return &KafkaProducer{client, logger}
}

// NewKafkaConsumer creates a new KafkaConsumer instance
func NewKafkaConsumer(client *kafka.KafkaClient, logger logger.Interface) Consumer {
	return &KafkaConsumer{client, logger}
}

// PublishListCreatedEvent publishes a list created event to the Kafka topic
func (p *KafkaProducer) PublishListCreatedEvent(userId int, list *entity.List) {
	message := entity.ListCreatedEvent{
		UserId: userId,
		List:   *list,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.Errorf("failed to marshal %s message: %v", ListCreatedTopic, err)
		return
	}

	if err := p.Publish(ListCreatedTopic, msgBytes); err != nil {
		p.logger.Errorf("failed to publish event to %s topic: %v", ListCreatedTopic, err)
	}
}

// PublishListUpdatedEvent publishes an event to notify that a list was updated
func (p *KafkaProducer) PublishListUpdatedEvent(userId, listId int) {
	message := entity.ListUpdatedEvent{
		UserId: userId,
		ListId: listId,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.Errorf("failed to marshal %s message: %v", ListUpdatedTopic, err)
		return
	}

	if err := p.Publish(ListUpdatedTopic, msgBytes); err != nil {
		p.logger.Errorf("failed to publish event to %s topic: %v", ListUpdatedTopic, err)
	}
}

// PublishListDeletedEvent publishes an event to notify that a list was deleted
func (p *KafkaProducer) PublishListDeletedEvent(listId int) {
	message := entity.ListDeletedEvent{
		ListId: listId,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.Errorf("failed to marshal %s message: %v", ListDeletedTopic, err)
		return
	}

	if err := p.Publish(ListDeletedTopic, msgBytes); err != nil {
		p.logger.Errorf("failed to publish event to %s topic: %v", ListDeletedTopic, err)
	}
}

// PublishItemCreatedEvent publishes an event to notify that an item was created
func (p *KafkaProducer) PublishItemCreatedEvent(userId, listId int, item *entity.Item) {
	message := entity.ItemCreatedEvent{
		UserId: userId,
		ListId: listId,
		Item:   *item,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.Errorf("failed to marshal %s message: %v", ItemCreatedTopic, err)
		return
	}

	if err := p.Publish(ItemCreatedTopic, msgBytes); err != nil {
		p.logger.Errorf("failed to publish event to %s topic: %v", ItemCreatedTopic, err)
	}
}

// PublishItemUpdatedEvent publishes an event to notify that an item was updated
func (p *KafkaProducer) PublishItemUpdatedEvent(userId, listId, itemId int) {
	message := entity.ItemUpdatedEvent{
		UserId: userId,
		ListId: listId,
		ItemId: itemId,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.Errorf("failed to marshal %s message: %v", ItemUpdatedTopic, err)
		return
	}

	if err := p.Publish(ItemUpdatedTopic, msgBytes); err != nil {
		p.logger.Errorf("failed to publish event to %s topic: %v", ItemUpdatedTopic, err)
	}
}

// PublishItemDeletedEvent publishes an event to notify that an item was deleted
func (p *KafkaProducer) PublishItemDeletedEvent(itemId int) {
	message := entity.ItemDeletedEvent{
		ItemId: itemId,
	}

	msgBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.Errorf("failed to marshal %s message: %v", ItemDeletedTopic, err)
		return
	}

	if err := p.Publish(ItemDeletedTopic, msgBytes); err != nil {
		p.logger.Errorf("failed to publish event to %s topic: %v", ItemDeletedTopic, err)
	}
}

// Publish sends a message to the specified Kafka topic
func (p *KafkaProducer) Publish(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err := p.client.Producer.SendMessage(msg)
	if err != nil {
		return err
	}

	return nil
}

// Subscribe subscribes to a Kafka topic and handles incoming messages using the provided handler function
func (p *KafkaConsumer) Subscribe(topic string, handler func(message []byte) error) error {
	partitionList, err := p.client.Consumer.Partitions(topic)
	if err != nil {
		return err
	}

	for _, partition := range partitionList {
		pc, err := p.client.Consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return err
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for message := range pc.Messages() {
				if err := handler(message.Value); err != nil {
					p.logger.Errorf("error handling message from topic %s: %v", topic, err)
				}
			}
		}(pc)
	}

	return nil
}
