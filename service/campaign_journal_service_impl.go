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
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CampaignJournalServiceImpl struct {
	campaignJournalRepo repository.CampaignJournalRepository
	campaignRepo        repository.CampaignRepository
	clickEventRepo      repository.ClickEventRepository
	conversionEventRepo repository.ConversionEventRepository
	redisClient         redis.Client
}

func NewCampaignJournalService(
	campaignJournalRepo repository.CampaignJournalRepository,
	campaignRepo repository.CampaignRepository,
	clickEventRepo repository.ClickEventRepository,
	conversionEventRepo repository.ConversionEventRepository,
	redisClient redis.Client,
) CampaignJournalService {
	return &CampaignJournalServiceImpl{
		campaignJournalRepo: campaignJournalRepo,
		campaignRepo:        campaignRepo,
		clickEventRepo:      clickEventRepo,
		conversionEventRepo: conversionEventRepo,
		redisClient:         redisClient,
	}
}

func (s *CampaignJournalServiceImpl) CalculateYesterdayMetrics(ctx context.Context) error {
	yesterday := time.Now().AddDate(0, 0, -1)
	dateStr := yesterday.Format("2006-01-02")

	log.Printf("Calculating metrics for date: %s", dateStr)

	campaignIDs, err := s.campaignRepo.GetDistinctCampaignIDsFromClickEvents(ctx, dateStr)
	if err != nil {
		return fmt.Errorf("failed to get campaign IDs from click events: %w", err)
	}

	log.Printf("Found %d campaigns with click events on %s", len(campaignIDs), dateStr)

	for _, campaignID := range campaignIDs {
		if err := s.processCampaignMetrics(ctx, campaignID, yesterday, dateStr); err != nil {
			log.Printf("Failed to process metrics for campaign %s: %v", campaignID.String(), err)
			continue
		}
	}

	return nil
}

func (s *CampaignJournalServiceImpl) processCampaignMetrics(ctx context.Context, campaignID uuid.UUID, date time.Time, dateStr string) error {
	if err := s.ensureCampaignExists(ctx, campaignID); err != nil {
		return fmt.Errorf("failed to ensure campaign exists: %w", err)
	}

	clickCount, err := s.getClickCountFromRedis(ctx, campaignID, dateStr)
	if err != nil {
		log.Printf("Failed to get click count from Redis for campaign %s: %v", campaignID.String(), err)
		clickCount = 0
	}

	conversionCount, err := s.getConversionCountFromRedis(ctx, campaignID, dateStr)
	if err != nil {
		log.Printf("Failed to get conversion count from Redis for campaign %s: %v", campaignID.String(), err)
		conversionCount = 0
	}

	totalConversionValue, err := s.getTotalConversionValueFromDB(ctx, campaignID, dateStr)
	if err != nil {
		log.Printf("Failed to get total conversion value for campaign %s: %v", campaignID.String(), err)
		totalConversionValue = decimal.Zero
	}

	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	campaignJournal := &entity.CampaignJournal{
		CampaignID:           campaignID,
		Date:                 dateOnly,
		NumberOfClick:        &clickCount,
		NumberOfConversion:   &conversionCount,
		TotalConversionValue: &totalConversionValue,
		CreatedAt:            time.Now(),
	}

	existingJournal, err := s.campaignJournalRepo.GetByCampaignAndDate(ctx, campaignID, dateOnly)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing journal entry: %w", err)
	}

	if existingJournal != nil {
		existingJournal.NumberOfClick = &clickCount
		existingJournal.NumberOfConversion = &conversionCount
		existingJournal.TotalConversionValue = &totalConversionValue

		if err := s.campaignJournalRepo.Update(ctx, existingJournal); err != nil {
			return fmt.Errorf("failed to update campaign journal: %w", err)
		}
		log.Printf("Updated campaign journal for campaign %s on %s", campaignID.String(), dateStr)
	} else {
		if err := s.campaignJournalRepo.Create(ctx, campaignJournal); err != nil {
			return fmt.Errorf("failed to create campaign journal: %w", err)
		}
		log.Printf("Created campaign journal for campaign %s on %s", campaignID.String(), dateStr)
	}

	log.Printf("Campaign %s metrics - Clicks: %d, Conversions: %d, Total Value: %s",
		campaignID.String(), clickCount, conversionCount, totalConversionValue.String())

	return nil
}

func (s *CampaignJournalServiceImpl) ensureCampaignExists(ctx context.Context, campaignID uuid.UUID) error {
	_, err := s.campaignRepo.GetByID(ctx, campaignID)
	if err == nil {
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err // Unexpected error
	}

	campaign := &entity.Campaign{
		ID:        campaignID,
		Name:      fmt.Sprintf("Campaign %s", campaignID.String()[:8]),
		CreatedAt: time.Now(),
	}

	if err := s.campaignRepo.Create(ctx, campaign); err != nil {
		return fmt.Errorf("failed to create campaign: %w", err)
	}

	log.Printf("Created new campaign with ID: %s", campaignID.String())
	return nil
}

func (s *CampaignJournalServiceImpl) getClickCountFromRedis(ctx context.Context, campaignID uuid.UUID, date string) (int64, error) {
	key := fmt.Sprintf("click_count:%s:%s", campaignID.String(), date)
	countStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse click count: %w", err)
	}

	return count, nil
}

func (s *CampaignJournalServiceImpl) getConversionCountFromRedis(ctx context.Context, campaignID uuid.UUID, date string) (int64, error) {
	key := fmt.Sprintf("conversion_count:%s:%s", campaignID.String(), date)
	countStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse conversion count: %w", err)
	}

	return count, nil
}

func (s *CampaignJournalServiceImpl) getTotalConversionValueFromDB(ctx context.Context, campaignID uuid.UUID, date string) (decimal.Decimal, error) {
	totalValue, err := s.conversionEventRepo.GetTotalConversionValue(ctx, campaignID, date)

	if err != nil {
		return decimal.Zero, err
	}

	return totalValue, nil
}
