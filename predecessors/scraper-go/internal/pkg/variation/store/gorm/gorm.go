package gorm

import (
	variation "scraper-go/internal/pkg/variation"
	variationStore "scraper-go/internal/pkg/variation/store"

	"github.com/rs/zerolog/log"

	"gorm.io/gorm"
)

type store struct {
	database *gorm.DB
}

func New(
	database *gorm.DB,
) variationStore.Store {
	return &store{
		database: database,
	}
}

func (s *store) CreateOne(request variationStore.CreateOneRequest) (*variationStore.CreateOneResponse, error) {
	cO := request.Variation

	result := s.database.Create(&cO)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error creating product")
		return nil, result.Error
	}

	return &variationStore.CreateOneResponse{ID: cO.ID}, nil
}

func (s *store) FindOne(request variationStore.FindOneRequest) (*variationStore.FindOneResponse, error) {
	var p = variation.Variation{ID: request.ID}
	result := s.database.Model(variation.Variation{}).First(&p)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error finding product")
		return nil, result.Error
	}

	return &variationStore.FindOneResponse{Variation: p}, nil
}
