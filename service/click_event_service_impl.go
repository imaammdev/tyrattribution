package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	"tyrattribution/entity"
	"tyrattribution/redis"
	"tyrattribution/repository"

	"github.com/google/uuid"
)

type ClickEventServiceImpl struct {
	clickEventRepository repository.ClickEventRepository
	redisClient          redis.Client
}

func NewClickEventService(clickEventRepository repository.ClickEventRepository, redisClient redis.Client) ClickEventService {
	return &ClickEventServiceImpl{
		clickEventRepository: clickEventRepository,
		redisClient:          redisClient,
	}
}

func (s *ClickEventServiceImpl) CreateClickEvent(ctx context.Context, clickEvent *entity.ClickEvent) error {
	if err := s.clickEventRepository.Create(ctx, clickEvent); err != nil {
		return err
	}

	date := clickEvent.ClickDate.Format("2006-01-02")
	counterKey := fmt.Sprintf("click_count:%s:%s", clickEvent.CampaignID.String(), date)

	count, err := s.redisClient.Incr(ctx, counterKey)
	if err != nil {
		log.Printf("Failed to increment Redis counter for key %s: %v", counterKey, err)
	} else {
		if count == 1 {
			nextDay := time.Now().AddDate(0, 0, 1)
			endOfNextDay := time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 23, 59, 59, 0, nextDay.Location())
			secondsUntilExpiry := int(time.Until(endOfNextDay).Seconds())

			if expireErr := s.redisClient.Expire(ctx, counterKey, secondsUntilExpiry); expireErr != nil {
				log.Printf("Failed to set expiration for Redis key %s: %v", counterKey, expireErr)
			}
		}
		log.Printf("Incremented click counter for campaign %s on %s: %d", clickEvent.CampaignID.String(), date, count)
	}

	return nil
}

func (s *ClickEventServiceImpl) GetClickCountByCampaign(ctx context.Context, campaignID uuid.UUID, date time.Time) (int64, error) {
	dateStr := date.Format("2006-01-02")
	counterKey := fmt.Sprintf("click_count:%s:%s", campaignID.String(), dateStr)

	countStr, err := s.redisClient.Get(ctx, counterKey)
	if err != nil {
		return 0, nil
	}

	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		log.Printf("Failed to parse Redis counter value for key %s: %v", counterKey, err)
		return 0, nil
	}

	return count, nil
}

func (s *ClickEventServiceImpl) GetClickEventsByCampaignUserSourceWithinTimeWindow(ctx context.Context, campaignID uuid.UUID, userID uuid.UUID, source string, clickDate time.Time, timeWindowHours int) (*entity.ClickEvent, error) {
	return s.clickEventRepository.GetClickEventsByCampaignUserSourceWithinTimeWindow(ctx, campaignID, userID, source, clickDate, timeWindowHours)
}
