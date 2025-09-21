package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"tyrattribution/config"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

type ClickEventPublisher struct {
	producer sarama.SyncProducer
	topic    string
}

type ClickEvent struct {
	ClickID    uuid.UUID `json:"click_id"`
	CampaignID uuid.UUID `json:"campaign_id"`
	UserID     uuid.UUID `json:"user_id"`
	ClickDate  time.Time `json:"click_date"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewClickEventPublisher(cfg *config.Config) (*ClickEventPublisher, error) {
	brokerURL := cfg.KafkaUrl
	topic := cfg.KafkaClickTopic

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &ClickEventPublisher{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *ClickEventPublisher) PublishClickEvent(event ClickEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.ClickID.String()),
		Value: sarama.ByteEncoder(eventJSON),
	}

	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Click event published to partition %d at offset %d", partition, offset)
	return nil
}

func (p *ClickEventPublisher) Close() error {
	return p.producer.Close()
}
