package service

import (
	"context"
	"time"
	"tyrattribution/entity"

	"github.com/google/uuid"
)

type ClickEventService interface {
	CreateClickEvent(ctx context.Context, clickEvent *entity.ClickEvent) error
	GetClickCountByCampaign(ctx context.Context, campaignID uuid.UUID, date time.Time) (int64, error)
	GetClickEventsByCampaignUserSourceWithinTimeWindow(ctx context.Context, campaignID uuid.UUID, userID uuid.UUID, source string, clickDate time.Time, timeWindowHours int) (*entity.ClickEvent, error)
}