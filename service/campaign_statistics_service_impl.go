package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"tyrattribution/entity"
	"tyrattribution/redis"
	"tyrattribution/repository"
)

type CampaignStatisticsServiceImpl struct {
	campaignJournalRepo repository.CampaignJournalRepository
	redisClient         redis.Client
	db                  *gorm.DB
}

func NewCampaignStatisticsService(
	campaignJournalRepo repository.CampaignJournalRepository,
	redisClient redis.Client,
	db *gorm.DB,
) CampaignStatisticsService {
	return &CampaignStatisticsServiceImpl{
		campaignJournalRepo: campaignJournalRepo,
		redisClient:         redisClient,
		db:                  db,
	}
}

func (s *CampaignStatisticsServiceImpl) GetCampaignStatistics(ctx context.Context, campaignID uuid.UUID, groupBy string) (*CampaignStatisticsResponse, error) {
	var groupByType GroupBy
	switch groupBy {
	case string(GroupByDaily):
		groupByType = GroupByDaily
	case string(GroupByWeekly):
		groupByType = GroupByWeekly
	case string(GroupByMonthly):
		groupByType = GroupByMonthly
	default:
		return nil, fmt.Errorf("invalid groupBy parameter: %s. Must be daily, weekly, or monthly", groupBy)
	}

	historicalData, err := s.getHistoricalData(ctx, campaignID, groupByType)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	var todayData *CampaignStatisticsDataItem
	// Only include today's data for daily reports
	if groupByType == GroupByDaily {
		todayData, err = s.getTodayData(ctx, campaignID)
		if err != nil {
			log.Printf("Failed to get today's data from Redis: %v", err)
			todayData = nil
		}
	}

	combinedData := s.combineData(historicalData, todayData, groupByType)

	return &CampaignStatisticsResponse{
		CampaignID: campaignID.String(),
		GroupBy:    groupBy,
		Data:       combinedData,
	}, nil
}

func (s *CampaignStatisticsServiceImpl) getHistoricalData(ctx context.Context, campaignID uuid.UUID, groupBy GroupBy) ([]CampaignStatisticsDataItem, error) {
	var results []CampaignStatisticsDataItem

	var rows *gorm.DB

	switch groupBy {
	case GroupByDaily:
		rows = s.db.WithContext(ctx).
			Model(&entity.CampaignJournal{}).
			Select(`
				date as period,
				COALESCE(number_of_click, 0) as total_clicks,
				COALESCE(number_of_conversion, 0) as total_conversions,
				COALESCE(total_conversion_value, 0) as total_value
			`).
			Where("campaign_id = ? AND date < ?", campaignID, time.Now().Format("2006-01-02")).
			Order("date DESC").
			Limit(30) // Last 30 days

	case GroupByWeekly:
		rows = s.db.WithContext(ctx).
			Model(&entity.CampaignJournal{}).
			Select(`
				TO_CHAR(DATE_TRUNC('week', date), 'YYYY-MM-DD') as period,
				SUM(COALESCE(number_of_click, 0)) as total_clicks,
				SUM(COALESCE(number_of_conversion, 0)) as total_conversions,
				SUM(COALESCE(total_conversion_value, 0)) as total_value
			`).
			Where("campaign_id = ? AND date < ? AND date IS NOT NULL", campaignID, time.Now().Format("2006-01-02")).
			Group("DATE_TRUNC('week', date)").
			Having("DATE_TRUNC('week', date) IS NOT NULL").
			Order("DATE_TRUNC('week', date) DESC").
			Limit(12) // Last 12 weeks

	case GroupByMonthly:
		rows = s.db.WithContext(ctx).
			Model(&entity.CampaignJournal{}).
			Select(`
				TO_CHAR(DATE_TRUNC('month', date), 'YYYY-MM') as period,
				SUM(COALESCE(number_of_click, 0)) as total_clicks,
				SUM(COALESCE(number_of_conversion, 0)) as total_conversions,
				SUM(COALESCE(total_conversion_value, 0)) as total_value
			`).
			Where("campaign_id = ? AND date < ? AND date IS NOT NULL", campaignID, time.Now().Format("2006-01-02")).
			Group("DATE_TRUNC('month', date)").
			Having("DATE_TRUNC('month', date) IS NOT NULL").
			Order("DATE_TRUNC('month', date) DESC").
			Limit(12) // Last 12 months
	}

	type QueryResult struct {
		Period           *string         `json:"period"`
		TotalClicks      int64           `json:"total_clicks"`
		TotalConversions int64           `json:"total_conversions"`
		TotalValue       decimal.Decimal `json:"total_value"`
	}

	var queryResults []QueryResult
	if err := rows.Scan(&queryResults).Error; err != nil {
		return nil, err
	}

	for _, result := range queryResults {
		var periodStr string

		if result.Period == nil {
			log.Printf("Warning: Period is nil for result with clicks: %d, conversions: %d", result.TotalClicks, result.TotalConversions)
			continue // Skip entries with nil periods
		}

		periodStr = *result.Period

		conversionRate := s.calculateConversionRate(result.TotalClicks, result.TotalConversions)

		results = append(results, CampaignStatisticsDataItem{
			Period:           periodStr,
			TotalClicks:      result.TotalClicks,
			TotalConversions: result.TotalConversions,
			TotalValue:       result.TotalValue,
			ConversionRate:   conversionRate,
		})
	}

	return results, nil
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

	var totalValue decimal.Decimal
	err = s.db.WithContext(ctx).
		Model(&entity.ConversionEvent{}).
		Select("COALESCE(SUM(value), 0)").
		Where("campaign_id = ? AND DATE(conversion_date) = ? AND click_id IS NOT NULL", campaignID, today).
		Scan(&totalValue).Error

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

func (s *CampaignStatisticsServiceImpl) combineData(historical []CampaignStatisticsDataItem, today *CampaignStatisticsDataItem, groupBy GroupBy) []CampaignStatisticsDataItem {
	if today == nil || (today.TotalClicks == 0 && today.TotalConversions == 0) {
		return historical
	}

	var todayPeriod string
	switch groupBy {
	case GroupByDaily:
		todayPeriod = time.Now().Format("2006-01-02")
	case GroupByWeekly:
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
		todayPeriod = weekStart.Format("2006-01-02")
	case GroupByMonthly:
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

func (s *CampaignStatisticsServiceImpl) formatPeriod(period time.Time, groupBy GroupBy) string {
	switch groupBy {
	case GroupByDaily:
		return period.Format("2006-01-02")
	case GroupByWeekly:
		return period.Format("2006-01-02") // Week start date
	case GroupByMonthly:
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