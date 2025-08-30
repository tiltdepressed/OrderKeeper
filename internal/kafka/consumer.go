// Package kafka
package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"orderkeeper/internal/models"
	"orderkeeper/internal/service"
	"strings"

	kafka "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	svc    service.OrderService
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	svc service.OrderService,
) *Consumer {
	log.Printf("Initializing Kafka consumer with brokers: %v, topic: %s, groupID: %s", brokers, topic, groupID)
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		svc: svc,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	defer c.reader.Close()
	log.Println("Kafka consumer is running and waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer due to context cancellation")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			log.Printf("Message received on topic %s, partition %d, offset %d", msg.Topic, msg.Partition, msg.Offset)

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Failed to unmarshal order: %v. Message: %s", err, string(msg.Value))
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Printf("Failed to commit bad message: %v", err)
				}
				continue
			}

			if err := c.svc.CreateOrder(order); err != nil {
				log.Printf("Failed to save order '%s': %v", order.OrderUID, err)
				continue
			}

			log.Printf("Order '%s' processed and saved successfully.", order.OrderUID)

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message for order '%s': %v", order.OrderUID, err)
			}
		}
	}
}

// InitKafkaConsumer инициализирует консьюмер из переданных параметров.
func InitKafkaConsumer(brokersStr, topic, groupID string, orderService service.OrderService) (*Consumer, error) {
	if brokersStr == "" {
		return nil, errors.New("kafka brokers string is not set")
	}
	if topic == "" {
		return nil, errors.New("kafka topic is not set")
	}
	if groupID == "" {
		return nil, errors.New("kafka groupID is not set")
	}

	brokers := strings.Split(brokersStr, ",")

	return NewConsumer(
		brokers,
		topic,
		groupID,
		orderService,
	), nil
}
