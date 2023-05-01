package variation

import "time"

type Variation struct {
	ID string `gorm:"id"`

	Identifier string `gorm:"identifier"` // could be SKU or VariationId

	DateAdded   time.Time `gorm:"date_added"`
	DateUpdated time.Time `gorm:"date_updated"`
}
