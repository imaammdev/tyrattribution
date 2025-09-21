package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"tyrattribution/entity"
)

type campaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB) CampaignRepository {
	return &campaignRepository{
		db: db,
	}
}

func (r *campaignRepository) Create(ctx context.Context, campaign *entity.Campaign) error {
	return r.db.WithContext(ctx).Create(campaign).Error
}

func (r *campaignRepository) GetByID(ctx context.Context, campaignID uuid.UUID) (*entity.Campaign, error) {
	var campaign entity.Campaign

	err := r.db.WithContext(ctx).
		Where("id = ?", campaignID).
		First(&campaign).Error

	if err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (r *campaignRepository) GetDistinctCampaignIDsFromClickEvents(ctx context.Context, date string) ([]uuid.UUID, error) {
	var campaignIDs []uuid.UUID

	err := r.db.WithContext(ctx).
		Model(&entity.ClickEvent{}).
		Select("DISTINCT campaign_id").
		Where("DATE(click_date) = ?", date).
		Pluck("campaign_id", &campaignIDs).Error

	return campaignIDs, err
}