package site

import "time"

type Site struct {
	ID   string `gorm:"id"`
	Url  string `gorm:"url"`
	Name string `gorm:"name"`

	Technology string `gorm:"technology"`

	DateAdded   time.Time `gorm:"date_added"`
	DateUpdated time.Time `gorm:"date_updated"`
}
