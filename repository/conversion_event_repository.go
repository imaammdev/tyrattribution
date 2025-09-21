package repository

import (
	"context"
	"tyrattribution/entity"
)

type ConversionEventRepository interface {
	Create(ctx context.Context, conversionEvent *entity.ConversionEvent) error
	Update(ctx context.Context, conversionEvent *entity.ConversionEvent) error
}
