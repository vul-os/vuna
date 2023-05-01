package store

import (
	"scraper-go/internal/pkg/site"

	"github.com/google/uuid"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
}

type CreateOneRequest struct {
	Site site.Site
}

type CreateOneResponse struct {
	ID uuid.UUID
}

type FindOneRequest struct {
	ID uuid.UUID
}

type FindOneResponse struct {
	Site site.Site
}
