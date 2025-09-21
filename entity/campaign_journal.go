package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CampaignJournal struct {
	CampaignJournalID    uuid.UUID        `json:"campaign_journal_id" gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:campaign_journal_id"`
	CampaignID           uuid.UUID        `json:"campaign_id" gorm:"type:uuid;not null;column:campaign_id"`
	Date                 time.Time        `json:"date" gorm:"type:date;not null;column:date"`
	NumberOfClick        *int64           `json:"number_of_click" gorm:"type:bigint;column:number_of_click"`
	NumberOfConversion   *int64           `json:"number_of_conversion" gorm:"type:bigint;column:number_of_conversion"`
	TotalConversionValue *decimal.Decimal `json:"total_conversion_value" gorm:"type:decimal(10,2);column:total_conversion_value"`
	CreatedAt            time.Time        `json:"created_at" gorm:"autoCreateTime;column:created_at"`
}

func (CampaignJournal) TableName() string {
	return "campaign_journal"
}
