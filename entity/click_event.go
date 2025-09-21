package entity

import (
	"time"

	"github.com/google/uuid"
)

type ClickEvent struct {
	ClickID    uuid.UUID `json:"click_id" gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:click_id"`
	CampaignID uuid.UUID `json:"campaign_id" gorm:"type:uuid;not null;column:campaign_id;index:idx_click_event_composite"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null;column:user_id;index:idx_click_event_composite"`
	ClickDate  time.Time `json:"click_date" gorm:"not null;column:click_date;index:idx_click_event_composite"`
	Source     string    `json:"source" gorm:"type:varchar(255);not null;column:source;index:idx_click_event_composite"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime;column:created_at"`
}

func (ClickEvent) TableName() string {
	return "click_event"
}
