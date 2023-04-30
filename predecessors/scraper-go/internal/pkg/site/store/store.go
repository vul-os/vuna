package store

import (
	"scraper-go/internal/pkg/site"
)

type Store interface {
	CreateOne(CreateOneRequest) (*CreateOneResponse, error)
	FindOne(FindOneRequest) (*FindOneResponse, error)
}

type CreateOneRequest struct {
	Site site.Site
}

type CreateOneResponse struct {
	ID string
}

type FindOneRequest struct {
	ID string
}

type FindOneResponse struct {
	Site site.Site
}
