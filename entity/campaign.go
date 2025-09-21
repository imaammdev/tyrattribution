package entity

import (
	"time"

	"github.com/google/uuid"
)

type Campaign struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:id"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null;column:name"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;column:created_at"`
}

func (Campaign) TableName() string {
	return "campaign"
}
