package gorm

import (
	site "scraper-go/internal/pkg/site"
	siteStore "scraper-go/internal/pkg/site/store"

	"github.com/rs/zerolog/log"

	"gorm.io/gorm"
)

type store struct {
	database *gorm.DB
}

func New(
	database *gorm.DB,
) siteStore.Store {
	return &store{
		database: database,
	}
}

func (s *store) CreateOne(request siteStore.CreateOneRequest) (*siteStore.CreateOneResponse, error) {
	cO := request.Site

	result := s.database.Create(&cO)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error creating product")
		return nil, result.Error
	}

	return &siteStore.CreateOneResponse{ID: cO.ID}, nil
}

func (s *store) FindOne(request siteStore.FindOneRequest) (*siteStore.FindOneResponse, error) {
	var p = site.Site{ID: request.ID}
	result := s.database.Model(site.Site{}).First(&p)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error finding product")
		return nil, result.Error
	}

	return &siteStore.FindOneResponse{Site: p}, nil
}
