package repository

import (
	"context"
	"log"
	"time"

	"tyrattribution/entity"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CampaignStatisticsRepositoryImpl struct {
	db *gorm.DB
}

func NewCampaignStatisticsRepository(db *gorm.DB) CampaignStatisticsRepository {
	return &CampaignStatisticsRepositoryImpl{
		db: db,
	}
}

func (r *CampaignStatisticsRepositoryImpl) GetHistoricalData(ctx context.Context, campaignID uuid.UUID, groupBy GroupBy) ([]CampaignStatisticsData, error) {
	var results []CampaignStatisticsData

	var rows *gorm.DB

	switch groupBy {
	case GroupByDaily:
		rows = r.db.WithContext(ctx).
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
		rows = r.db.WithContext(ctx).
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
			Order("DATE_TRUNC('week', date) DESC")

	case GroupByMonthly:
		rows = r.db.WithContext(ctx).
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

		results = append(results, CampaignStatisticsData{
			Period:           periodStr,
			TotalClicks:      result.TotalClicks,
			TotalConversions: result.TotalConversions,
			TotalValue:       result.TotalValue,
		})
	}

	return results, nil
}

func (r *CampaignStatisticsRepositoryImpl) GetTodayConversionValue(ctx context.Context, campaignID uuid.UUID, date time.Time) (decimal.Decimal, error) {
	var totalValue decimal.Decimal
	dateStr := date.Format("2006-01-02")

	err := r.db.WithContext(ctx).
		Model(&entity.ConversionEvent{}).
		Select("COALESCE(SUM(value), 0)").
		Where("campaign_id = ? AND DATE(conversion_date) = ? AND click_id IS NOT NULL", campaignID, dateStr).
		Scan(&totalValue).Error

	if err != nil {
		log.Printf("Failed to get conversion value for date %s: %v", dateStr, err)
		return decimal.Zero, err
	}

	return totalValue, nil
}
