package repository

import (
	"context"
	"tyrattribution/entity"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ConversionEventRepository interface {
	Create(ctx context.Context, conversionEvent *entity.ConversionEvent) error
	Update(ctx context.Context, conversionEvent *entity.ConversionEvent) error
	GetTotalConversionValue(ctx context.Context, campaignID uuid.UUID, date string) (decimal.Decimal, error)
}
