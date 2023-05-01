package store

import (
	"scraper-go/internal/pkg/variation"

	"github.com/google/uuid"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
}

type CreateOneRequest struct {
	Variation variation.Variation
}

type CreateOneResponse struct {
	ID uuid.UUID
}

type FindOneRequest struct {
	ID uuid.UUID
}

type FindOneResponse struct {
	Variation variation.Variation
}
