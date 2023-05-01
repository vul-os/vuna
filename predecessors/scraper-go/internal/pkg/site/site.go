package site

import (
	"time"

	"github.com/google/uuid"
)

type Site struct {
	ID   uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	Url  string `gorm:"url"`
	Name string `gorm:"name"`

	Technology string `gorm:"technology"`

	DateAdded   time.Time `gorm:"date_added"`
	DateUpdated time.Time `gorm:"date_updated"`
}
