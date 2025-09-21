package repository

import (
	"context"

	"tyrattribution/entity"

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
