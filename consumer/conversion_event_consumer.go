package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"tyrattribution/config"
	"tyrattribution/entity"
	"tyrattribution/publisher"
	"tyrattribution/service"

	"github.com/IBM/sarama"
)

type ConversionEventConsumer struct {
	consumer sarama.ConsumerGroup
	service  service.ConversionEventService
	topic    string
}

type ConversionEventMessage = publisher.ConversionEvent

func NewConversionEventConsumer(cfg *config.Config, svc service.ConversionEventService) (*ConversionEventConsumer, error) {
	brokerURL := cfg.KafkaUrl
	topic := cfg.KafkaConversionTopic
	groupID := "tyr"

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup([]string{brokerURL}, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &ConversionEventConsumer{
		consumer: consumer,
		service:  svc,
		topic:    topic,
	}, nil
}

func (c *ConversionEventConsumer) Start(ctx context.Context) error {
	handler := &conversionEventHandler{service: c.service}

	for {
		select {
		case <-ctx.Done():
			log.Println("Conversion event consumer context cancelled")
			return nil
		case err := <-c.consumer.Errors():
			log.Printf("Consumer error: %v", err)
		default:
			err := c.consumer.Consume(ctx, []string{c.topic}, handler)
			if err != nil {
				log.Printf("Error consuming messages: %v", err)
				return err
			}
		}
	}
}

func (c *ConversionEventConsumer) Close() error {
	return c.consumer.Close()
}

type conversionEventHandler struct {
	service service.ConversionEventService
}

func (h *conversionEventHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *conversionEventHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *conversionEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var eventMsg ConversionEventMessage
			if err := json.Unmarshal(message.Value, &eventMsg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			if err := h.saveConversionEventToDB(eventMsg); err != nil {
				log.Printf("Error saving conversion event to database: %v", err)
			}
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *conversionEventHandler) saveConversionEventToDB(eventMsg ConversionEventMessage) error {
	conversionEvent := &entity.ConversionEvent{
		ConversionID:   eventMsg.ConversionID,
		UserID:         eventMsg.UserID,
		CampaignID:     eventMsg.CampaignID,
		ConversionDate: eventMsg.ConversionDate,
		Value:          eventMsg.Value,
		Type:           eventMsg.Type,
		Source:         eventMsg.Source,
		CreatedAt:      eventMsg.CreatedAt,
	}

	return h.service.CreateConversionEvent(context.Background(), conversionEvent)
}

func StartConversionEventConsumer(ctx context.Context, cfg *config.Config, svc service.ConversionEventService) {
	consumer, err := NewConversionEventConsumer(cfg, svc)
	if err != nil {
		log.Fatalf("Failed to create conversion event consumer: %v", err)
	}

	log.Println("Starting conversion event consumer")
	if err := consumer.Start(ctx); err != nil {
		log.Printf("Conversion event consumer error: %v", err)
	}

	if err := consumer.Close(); err != nil {
		log.Printf("Error closing conversion event consumer: %v", err)
	}
}
