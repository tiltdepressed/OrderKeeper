// Package kafka
package kafka

import (
	"context"
	"encoding/json"
	"log"
	"orderkeeper/internal/models"
	"orderkeeper/internal/service"
	"os"

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

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Kafka consumer due to context cancellation")
			return
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Failed to unmarshal order: %v", err)
				continue
			}

			if err := c.svc.CreateOrder(order); err != nil {
				log.Printf("Failed to save order: %v", err)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}
}

func InitKafkaConsumer(orderService service.OrderService) (*Consumer, error) {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "orders"
	}
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "order-service-group"
	}

	return NewConsumer(
		[]string{kafkaBrokers},
		kafkaTopic,
		kafkaGroupID,
		orderService,
	), nil
}
