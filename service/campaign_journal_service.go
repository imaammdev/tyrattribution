package service

import (
	"context"
)

type CampaignJournalService interface {
	CalculateYesterdayMetrics(ctx context.Context) error
}