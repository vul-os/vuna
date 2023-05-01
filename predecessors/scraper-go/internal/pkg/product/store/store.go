package store

import (
	"scraper-go/internal/pkg/product"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
	UpsertOne(UpsertOneRequest) (*UpsertOneResponse, error)
	
}

type CreateOneRequest struct {
	Product product.Product
}

type CreateOneResponse struct {
	ID string
}

type UpsertOneRequest struct {
	Product product.Product
}

type UpsertOneResponse struct {
	ID string
}

type FindOneRequest struct {
	ID string
}

type FindOneResponse struct {
	Product product.Product
}


