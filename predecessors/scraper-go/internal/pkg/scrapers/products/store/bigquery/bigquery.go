package bigquery

import (
	"scraper-go/internal/pkg/scrapers/products"

	"cloud.google.com/go/bigquery"

	productsStore "scraper-go/internal/pkg/scrapers/products/store"
)

type store struct {
	client *bigquery.Client
}

func New(
	client *bigquery.Client,
) productsStore.Store {
	return &store{
		client: client,
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
