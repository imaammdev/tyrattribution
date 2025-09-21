package repository

import (
	"context"

	"github.com/google/uuid"
	"tyrattribution/entity"
)

type CampaignRepository interface {
	Create(ctx context.Context, campaign *entity.Campaign) error
	GetByID(ctx context.Context, campaignID uuid.UUID) (*entity.Campaign, error)
	GetDistinctCampaignIDsFromClickEvents(ctx context.Context, date string) ([]uuid.UUID, error)
}