package store

import (
	"scraper-go/internal/pkg/product"

	"github.com/google/uuid"
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
	ID uuid.UUID
}

type UpsertOneRequest struct {
	Product product.Product
}

type UpsertOneResponse struct {
	ID uuid.UUID
}

type FindOneRequest struct {
	ID uuid.UUID
}

type FindOneResponse struct {
	Product product.Product
}
