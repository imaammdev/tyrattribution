package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ConversionEvent struct {
	ConversionID   uuid.UUID        `json:"conversion_id" gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:conversion_id"`
	UserID         uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;column:user_id"`
	CampaignID     uuid.UUID        `json:"campaign_id" gorm:"type:uuid;not null;column:campaign_id;index"`
	ClickID        *uuid.UUID       `json:"click_id" gorm:"type:uuid;column:click_id"`
	ConversionDate time.Time        `json:"conversion_date" gorm:"not null;column:conversion_date"`
	Value          *decimal.Decimal `json:"value" gorm:"type:decimal(10,2);column:value"`
	Type           string           `json:"type" gorm:"type:varchar(255);not null;column:type"`
	Source         string           `json:"source" gorm:"type:varchar(255);not null;column:source"`
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime;column:created_at"`
}

func (ConversionEvent) TableName() string {
	return "conversion_event"
}
