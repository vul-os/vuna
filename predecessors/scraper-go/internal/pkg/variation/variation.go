package variation

import (
	"time"

	"github.com/google/uuid"
)

type Variation struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	ProductId uuid.UUID `json:"productId" gorm:"productId"`
	
	Identifier string `gorm:"identifier"` // could be SKU or VariationId

	DateAdded   time.Time `gorm:"date_added"`
	DateUpdated time.Time `gorm:"date_updated"`
}
