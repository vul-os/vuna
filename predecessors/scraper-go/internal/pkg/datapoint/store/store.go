package store

import (
	"scraper-go/internal/pkg/datapoint"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
}

type CreateOneRequest struct {
	Datapoint datapoint.DataPoint
}

type CreateOneResponse struct {
	ID string
}

type FindOneRequest struct {
	ID string
}

type FindOneResponse struct {
	Datapoint datapoint.DataPoint
}
