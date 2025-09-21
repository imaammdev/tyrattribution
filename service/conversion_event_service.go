package service

import (
	"context"
	"tyrattribution/entity"
)

type ConversionEventService interface {
	CreateConversionEvent(ctx context.Context, conversionEvent *entity.ConversionEvent) error
}