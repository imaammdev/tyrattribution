package repository

import (
	"context"

	"tyrattribution/entity"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type conversionEventRepository struct {
	db *gorm.DB
}

func NewConversionEventRepository(db *gorm.DB) ConversionEventRepository {
	return &conversionEventRepository{
		db: db,
	}
}

func (r *conversionEventRepository) Create(ctx context.Context, conversionEvent *entity.ConversionEvent) error {
	return r.db.WithContext(ctx).Create(conversionEvent).Error
}

func (r *conversionEventRepository) Update(ctx context.Context, conversionEvent *entity.ConversionEvent) error {
	return r.db.WithContext(ctx).Save(conversionEvent).Error
}

func (s *conversionEventRepository) GetTotalConversionValue(ctx context.Context, campaignID uuid.UUID, date string) (decimal.Decimal, error) {
	var totalValue decimal.Decimal

	err := s.db.WithContext(ctx).
		Model(&entity.ConversionEvent{}).
		Select("COALESCE(SUM(value), 0)").
		Where("campaign_id = ? AND DATE(conversion_date) = ? AND click_id IS NOT NULL", campaignID, date).
		Scan(&totalValue).Error

	if err != nil {
		return decimal.Zero, err
	}

	return totalValue, nil
}
