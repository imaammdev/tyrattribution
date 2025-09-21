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

type ClickEventConsumer struct {
	consumer sarama.ConsumerGroup
	service  service.ClickEventService
	topic    string
}

type ClickEventMessage = publisher.ClickEvent

func NewClickEventConsumer(cfg *config.Config, svc service.ClickEventService) (*ClickEventConsumer, error) {
	brokerURL := cfg.KafkaUrl
	topic := cfg.KafkaClickTopic
	groupID := "tyr"

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup([]string{brokerURL}, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &ClickEventConsumer{
		consumer: consumer,
		service:  svc,
		topic:    topic,
	}, nil
}

func (c *ClickEventConsumer) Start(ctx context.Context) error {
	handler := &clickEventHandler{service: c.service}

	for {
		select {
		case <-ctx.Done():
			log.Println("Click event consumer context cancelled")
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

func (c *ClickEventConsumer) Close() error {
	return c.consumer.Close()
}

type clickEventHandler struct {
	service service.ClickEventService
}

func (h *clickEventHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *clickEventHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *clickEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var eventMsg ClickEventMessage
			if err := json.Unmarshal(message.Value, &eventMsg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				session.MarkMessage(message, "")
				continue
			}

			if err := h.saveClickEventToDB(eventMsg); err != nil {
				log.Printf("Error saving click event to database: %v", err)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func (h *clickEventHandler) saveClickEventToDB(eventMsg ClickEventMessage) error {
	clickEvent := &entity.ClickEvent{
		ClickID:    eventMsg.ClickID,
		CampaignID: eventMsg.CampaignID,
		UserID:     eventMsg.UserID,
		ClickDate:  eventMsg.ClickDate,
		Source:     eventMsg.Source,
		CreatedAt:  eventMsg.CreatedAt,
	}

	return h.service.CreateClickEvent(context.Background(), clickEvent)
}

func StartClickEventConsumer(ctx context.Context, cfg *config.Config, svc service.ClickEventService) {
	consumer, err := NewClickEventConsumer(cfg, svc)
	if err != nil {
		log.Fatalf("Failed to create click event consumer: %v", err)
	}

	log.Println("Starting click event consumer")
	if err := consumer.Start(ctx); err != nil {
		log.Printf("Click event consumer error: %v", err)
	}

	if err := consumer.Close(); err != nil {
		log.Printf("Error closing click event consumer: %v", err)
	}
}
