package service

import (
	"context"
	"fmt"
	"log"
	"time"
	"tyrattribution/config"
	"tyrattribution/entity"
	"tyrattribution/redis"
	"tyrattribution/repository"
)

type ConversionEventServiceImpl struct {
	conversionEventRepository repository.ConversionEventRepository
	clickEventService         ClickEventService
	redisClient               redis.Client
	config                    *config.Config
}

func NewConversionEventService(conversionEventRepository repository.ConversionEventRepository, clickEventService ClickEventService, redisClient redis.Client, cfg *config.Config) ConversionEventService {
	return &ConversionEventServiceImpl{
		conversionEventRepository: conversionEventRepository,
		clickEventService:         clickEventService,
		redisClient:               redisClient,
		config:                    cfg,
	}
}

func (s *ConversionEventServiceImpl) CreateConversionEvent(ctx context.Context, conversionEvent *entity.ConversionEvent) error {
	if err := s.conversionEventRepository.Create(ctx, conversionEvent); err != nil {
		return err
	}

	timeWindowHours := s.config.ClickEventTimeWindowHours
	matchedClick, err := s.clickEventService.GetClickEventsByCampaignUserSourceWithinTimeWindow(
		ctx,
		conversionEvent.CampaignID,
		conversionEvent.UserID,
		conversionEvent.Source,
		conversionEvent.ConversionDate,
		timeWindowHours,
	)

	if err != nil {
		log.Printf("Error checking for matched click event: %v", err)
		return nil
	}

	if matchedClick != nil {
		conversionEvent.ClickID = &matchedClick.ClickID

		if err := s.conversionEventRepository.Update(ctx, conversionEvent); err != nil {
			log.Printf("Failed to update conversion event with ClickID: %v", err)
		} else {
			log.Printf("Attributed conversion %s to click %s", conversionEvent.ConversionID.String(), matchedClick.ClickID.String())
			s.incrementConversionCounter(ctx, conversionEvent)
		}
	} else {
		log.Printf("No matching click event found for conversion %s within %d hour window", conversionEvent.ConversionID.String(), timeWindowHours)
	}

	return nil
}

func (s *ConversionEventServiceImpl) incrementConversionCounter(ctx context.Context, conversionEvent *entity.ConversionEvent) {
	date := conversionEvent.ConversionDate.Format("2006-01-02")
	counterKey := fmt.Sprintf("conversion_count:%s:%s", conversionEvent.CampaignID.String(), date)

	count, err := s.redisClient.Incr(ctx, counterKey)
	if err != nil {
		log.Printf("Failed to increment Redis conversion counter for key %s: %v", counterKey, err)
		return
	}

	if count == 1 {
		nextDay := time.Now().AddDate(0, 0, 1)
		endOfNextDay := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 23, 59, 59, 0, nextDay.Location())
		secondsUntilExpiry := int(time.Until(endOfNextDay).Seconds())

		if expireErr := s.redisClient.Expire(ctx, counterKey, secondsUntilExpiry); expireErr != nil {
			log.Printf("Failed to set expiration for Redis key %s: %v", counterKey, expireErr)
		}
	}

	log.Printf("Incremented conversion counter for campaign %s on %s: %d", conversionEvent.CampaignID.String(), date, count)
}