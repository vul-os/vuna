package orchestrator

import (
	"gorm.io/gorm"
)

func Orchestrator(db *gorm.DB, batchSize int) error {
	products, err := GetProducts(db, batchSize)
	if err != nil {
		return err
	}

	for _, product := range products {
		if err := CreateTaskForProduct(product); err != nil {
			return err
		}
	}

	return nil
}
