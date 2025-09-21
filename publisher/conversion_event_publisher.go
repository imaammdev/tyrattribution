package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ConversionEventPublisher struct {
	producer sarama.SyncProducer
	topic    string
}

type ConversionEvent struct {
	ConversionID   uuid.UUID        `json:"conversion_id"`
	UserID         uuid.UUID        `json:"user_id"`
	CampaignID     uuid.UUID        `json:"campaign_id"`
	ConversionDate time.Time        `json:"conversion_date"`
	Value          *decimal.Decimal `json:"value"`
	Type           string           `json:"type"`
	Source         string           `json:"source"`
	CreatedAt      time.Time        `json:"created_at"`
}

func NewConversionEventPublisher() (*ConversionEventPublisher, error) {
	brokerURL := os.Getenv("KAFKA_BROKER_URL")
	topic := os.Getenv("KAFKA_CONVERSION_EVENT_TOPIC")

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &ConversionEventPublisher{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *ConversionEventPublisher) PublishConversionEvent(event ConversionEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.ConversionID.String()),
		Value: sarama.ByteEncoder(eventJSON),
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Conversion event published to partition %d at offset %d", partition, offset)
	return nil
}

func (p *ConversionEventPublisher) Close() error {
	return p.producer.Close()
}