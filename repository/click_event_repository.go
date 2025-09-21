package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"tyrattribution/entity"
)

type ClickEventRepository interface {
	GetClickEventsByCampaignUserSourceWithinTimeWindow(
		ctx context.Context,
		campaignID uuid.UUID,
		userID uuid.UUID,
		source string,
		clickDate time.Time,
		timeWindowHours int,
	) (*entity.ClickEvent, error)

	Create(ctx context.Context, clickEvent *entity.ClickEvent) error
}