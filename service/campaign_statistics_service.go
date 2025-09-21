package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CampaignStatisticsService interface {
	GetCampaignStatistics(ctx context.Context, campaignID uuid.UUID, groupBy string) (*CampaignStatisticsResponse, error)
}


type CampaignStatisticsResponse struct {
	CampaignID string                       `json:"campaign_id"`
	GroupBy    string                       `json:"group_by"`
	Data       []CampaignStatisticsDataItem `json:"data"`
}

type CampaignStatisticsDataItem struct {
	Period           string          `json:"period"`
	TotalClicks      int64           `json:"total_clicks"`
	TotalConversions int64           `json:"total_conversions"`
	TotalValue       decimal.Decimal `json:"total_value"`
	ConversionRate   float64         `json:"conversion_rate"`
}
