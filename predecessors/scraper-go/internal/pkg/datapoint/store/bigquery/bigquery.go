package bigquery

import (

	"cloud.google.com/go/bigquery"

	dp "scraper-go/internal/pkg/datapoint"
	dpStore "scraper-go/internal/pkg/datapoint/store"
)

type store struct {
	client *bigquery.Client
}

func New(
	client *bigquery.Client,
) dpStore.Store {
	return &store{
		client: client,
	}
}

func (s *store) CreateOne(request dpStore.CreateOneRequest) (*dpStore.CreateOneResponse, error) {
	cO := request.Datapoint

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
	
	return &dpStore.CreateOneResponse{ID: cO.ID}, nil
}

func (s *store) FindOne(request dpStore.FindOneRequest) (*dpStore.FindOneResponse, error) {
	return &dpStore.FindOneResponse{Datapoint: dp.DataPoint{}}, nil
}
