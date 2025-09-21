package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CampaignStatisticsData struct {
	Period           string          `json:"period"`
	TotalClicks      int64           `json:"total_clicks"`
	TotalConversions int64           `json:"total_conversions"`
	TotalValue       decimal.Decimal `json:"total_value"`
}

type GroupBy string

const (
	GroupByDaily   GroupBy = "daily"
	GroupByWeekly  GroupBy = "weekly"
	GroupByMonthly GroupBy = "monthly"
)

type CampaignStatisticsRepository interface {
	GetHistoricalData(ctx context.Context, campaignID uuid.UUID, groupBy GroupBy) ([]CampaignStatisticsData, error)
	GetTodayConversionValue(ctx context.Context, campaignID uuid.UUID, date time.Time) (decimal.Decimal, error)
}