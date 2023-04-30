package gorm

import (
	"scraper-go/internal/pkg/scrapers/products"
	productsStore "scraper-go/internal/pkg/scrapers/products/store"

	"gorm.io/gorm"
)

type store struct {
	database *gorm.DB
}

func New(
	database *gorm.DB,
) productsStore.Store {
	return &store{
		database: database,
	}
}

func (s *store) CreateOne(request productsStore.CreateOneRequest) (*productsStore.CreateOneResponse, error) {
	cO := request.Product

	// inserter := s.client.Dataset(datasetID).Table(tableID).Inserter()
	// items := []*Item{
	// 	// Item implements the ValueSaver interface.
	// 	{Name: "Phred Phlyntstone", Age: 32},
	// 	{Name: "Wylma Phlyntstone", Age: 29},
	// }
	// if err := inserter.Put(ctx, items); err != nil {
	// 	log.Error().Err(result.Error).Msg("error creating dashboard")
	// 	return nil, result.Error
	// }
	return &productsStore.CreateOneResponse{ID: cO.ID}, nil
}

func (s *store) FindOne(request productsStore.FindOneRequest) (*productsStore.FindOneResponse, error) {
	return &productsStore.FindOneResponse{Product: products.Product{}}, nil

}