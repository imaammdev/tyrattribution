package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"tyrattribution/entity"
)

type clickEventRepository struct {
	db *gorm.DB
}

func NewClickEventRepository(db *gorm.DB) ClickEventRepository {
	return &clickEventRepository{
		db: db,
	}
}

func (r *clickEventRepository) GetClickEventsByCampaignUserSourceWithinTimeWindow(
	ctx context.Context,
	campaignID uuid.UUID,
	userID uuid.UUID,
	source string,
	clickDate time.Time,
	timeWindowHours int,
) (*entity.ClickEvent, error) {
	var clickEvent entity.ClickEvent

	startTime := clickDate.Add(-time.Duration(timeWindowHours) * time.Hour)
	endTime := clickDate.Add(time.Duration(timeWindowHours) * time.Hour)

	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND user_id = ? AND source = ? AND click_date BETWEEN ? AND ?",
			campaignID, userID, source, startTime, endTime).
		Order("click_date DESC").
		First(&clickEvent).Error

	if err != nil {
		return nil, err
	}

	return &clickEvent, nil
}

func (r *clickEventRepository) Create(ctx context.Context, clickEvent *entity.ClickEvent) error {
	return r.db.WithContext(ctx).Create(clickEvent).Error
}