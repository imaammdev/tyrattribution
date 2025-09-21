package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"tyrattribution/redis"
	"tyrattribution/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CampaignStatisticsServiceImpl struct {
	campaignJournalRepo repository.CampaignJournalRepository
	campaignStatsRepo   repository.CampaignStatisticsRepository
	redisClient         redis.Client
}

func NewCampaignStatisticsService(
	campaignJournalRepo repository.CampaignJournalRepository,
	campaignStatsRepo repository.CampaignStatisticsRepository,
	redisClient redis.Client,
) CampaignStatisticsService {
	return &CampaignStatisticsServiceImpl{
		campaignJournalRepo: campaignJournalRepo,
		campaignStatsRepo:   campaignStatsRepo,
		redisClient:         redisClient,
	}
}

func (s *CampaignStatisticsServiceImpl) GetCampaignStatistics(ctx context.Context, campaignID uuid.UUID, groupBy string) (*CampaignStatisticsResponse, error) {
	var groupByType repository.GroupBy
	switch groupBy {
	case string(repository.GroupByDaily):
		groupByType = repository.GroupByDaily
	case string(repository.GroupByWeekly):
		groupByType = repository.GroupByWeekly
	case string(repository.GroupByMonthly):
		groupByType = repository.GroupByMonthly
	default:
		return nil, fmt.Errorf("invalid groupBy parameter: %s. Must be daily, weekly, or monthly", groupBy)
	}

	historicalData, err := s.campaignStatsRepo.GetHistoricalData(ctx, campaignID, groupByType)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	var todayData *CampaignStatisticsDataItem
	// Only include today's data for daily reports
	if groupByType == repository.GroupByDaily {
		todayData, err = s.getTodayData(ctx, campaignID)
		if err != nil {
			log.Printf("Failed to get today's data from Redis: %v", err)
			todayData = nil
		}
	}

	// Convert repository data to service data
	serviceHistoricalData := s.convertToServiceData(historicalData)

	combinedData := s.combineData(serviceHistoricalData, todayData, groupByType)

	return &CampaignStatisticsResponse{
		CampaignID: campaignID.String(),
		GroupBy:    groupBy,
		Data:       combinedData,
	}, nil
}

func (s *CampaignStatisticsServiceImpl) getTodayData(ctx context.Context, campaignID uuid.UUID) (*CampaignStatisticsDataItem, error) {
	today := time.Now().Format("2006-01-02")

	clickKey := fmt.Sprintf("click_count:%s:%s", campaignID.String(), today)
	clickCountStr, err := s.redisClient.Get(ctx, clickKey)
	var clickCount int64 = 0
	if err == nil {
		if parsed, parseErr := strconv.ParseInt(clickCountStr, 10, 64); parseErr == nil {
			clickCount = parsed
		}
	}

	conversionKey := fmt.Sprintf("conversion_count:%s:%s", campaignID.String(), today)
	conversionCountStr, err := s.redisClient.Get(ctx, conversionKey)
	var conversionCount int64 = 0
	if err == nil {
		if parsed, parseErr := strconv.ParseInt(conversionCountStr, 10, 64); parseErr == nil {
			conversionCount = parsed
		}
	}

	totalValue, err := s.campaignStatsRepo.GetTodayConversionValue(ctx, campaignID, time.Now())
	if err != nil {
		log.Printf("Failed to get today's total conversion value: %v", err)
		totalValue = decimal.Zero
	}

	conversionRate := s.calculateConversionRate(clickCount, conversionCount)

	return &CampaignStatisticsDataItem{
		Period:           today,
		TotalClicks:      clickCount,
		TotalConversions: conversionCount,
		TotalValue:       totalValue,
		ConversionRate:   conversionRate,
	}, nil
}

func (s *CampaignStatisticsServiceImpl) convertToServiceData(repoData []repository.CampaignStatisticsData) []CampaignStatisticsDataItem {
	var result []CampaignStatisticsDataItem
	for _, data := range repoData {
		conversionRate := s.calculateConversionRate(data.TotalClicks, data.TotalConversions)
		result = append(result, CampaignStatisticsDataItem{
			Period:           data.Period,
			TotalClicks:      data.TotalClicks,
			TotalConversions: data.TotalConversions,
			TotalValue:       data.TotalValue,
			ConversionRate:   conversionRate,
		})
	}
	return result
}

func (s *CampaignStatisticsServiceImpl) combineData(historical []CampaignStatisticsDataItem, today *CampaignStatisticsDataItem, groupBy repository.GroupBy) []CampaignStatisticsDataItem {
	if today == nil || (today.TotalClicks == 0 && today.TotalConversions == 0) {
		return historical
	}

	var todayPeriod string
	switch groupBy {
	case repository.GroupByDaily:
		todayPeriod = time.Now().Format("2006-01-02")
	case repository.GroupByWeekly:
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
		todayPeriod = weekStart.Format("2006-01-02")
	case repository.GroupByMonthly:
		todayPeriod = time.Now().Format("2006-01")
	}

	found := false
	for i, item := range historical {
		if item.Period == todayPeriod {
			historical[i].TotalClicks += today.TotalClicks
			historical[i].TotalConversions += today.TotalConversions
			historical[i].TotalValue = historical[i].TotalValue.Add(today.TotalValue)
			historical[i].ConversionRate = s.calculateConversionRate(historical[i].TotalClicks, historical[i].TotalConversions)
			found = true
			break
		}
	}

	if !found {
		todayData := CampaignStatisticsDataItem{
			Period:           todayPeriod,
			TotalClicks:      today.TotalClicks,
			TotalConversions: today.TotalConversions,
			TotalValue:       today.TotalValue,
			ConversionRate:   s.calculateConversionRate(today.TotalClicks, today.TotalConversions),
		}

		historical = append([]CampaignStatisticsDataItem{todayData}, historical...)
	}

	return historical
}

func (s *CampaignStatisticsServiceImpl) formatPeriod(period time.Time, groupBy repository.GroupBy) string {
	switch groupBy {
	case repository.GroupByDaily:
		return period.Format("2006-01-02")
	case repository.GroupByWeekly:
		return period.Format("2006-01-02") // Week start date
	case repository.GroupByMonthly:
		return period.Format("2006-01")
	default:
		return period.Format("2006-01-02")
	}
}

func (s *CampaignStatisticsServiceImpl) calculateConversionRate(clicks, conversions int64) float64 {
	if clicks == 0 {
		return 0.0
	}
	return (float64(conversions) / float64(clicks)) * 100.0
}
