package repository

import (
	"context"
	"time"

	"tyrattribution/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type campaignJournalRepository struct {
	db *gorm.DB
}

func NewCampaignJournalRepository(db *gorm.DB) CampaignJournalRepository {
	return &campaignJournalRepository{
		db: db,
	}
}

func (r *campaignJournalRepository) Create(ctx context.Context, campaignJournal *entity.CampaignJournal) error {
	return r.db.WithContext(ctx).Create(campaignJournal).Error
}

func (r *campaignJournalRepository) GetByCampaignAndDate(ctx context.Context, campaignID uuid.UUID, date time.Time) (*entity.CampaignJournal, error) {
	var campaignJournal entity.CampaignJournal

	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND date = ?", campaignID, dateOnly).
		First(&campaignJournal).Error

	if err != nil {
		return nil, err
	}

	return &campaignJournal, nil
}

func (r *campaignJournalRepository) Update(ctx context.Context, campaignJournal *entity.CampaignJournal) error {
	return r.db.WithContext(ctx).Save(campaignJournal).Error
}
