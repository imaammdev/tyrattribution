package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"tyrattribution/entity"
)

type CampaignJournalRepository interface {
	Create(ctx context.Context, campaignJournal *entity.CampaignJournal) error
	GetByCampaignAndDate(ctx context.Context, campaignID uuid.UUID, date time.Time) (*entity.CampaignJournal, error)
	Update(ctx context.Context, campaignJournal *entity.CampaignJournal) error
}