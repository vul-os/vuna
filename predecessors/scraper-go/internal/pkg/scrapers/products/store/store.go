package store

import (
	"scraper-go/internal/pkg/scrapers/products"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
}

type CreateOneRequest struct {
	Product products.Product
}

type CreateOneResponse struct {
	ID string
}

type FindOneRequest struct {
	ID string
}

type FindOneResponse struct {
	Product products.Product
}
