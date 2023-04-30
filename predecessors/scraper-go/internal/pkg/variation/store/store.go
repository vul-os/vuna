package store

import (
	"scraper-go/internal/pkg/variation"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
}

type CreateOneRequest struct {
	Variation variation.Variation
}

type CreateOneResponse struct {
	ID string
}

type FindOneRequest struct {
	ID string
}

type FindOneResponse struct {
	Variation variation.Variation
}
