package gorm

import (
	product "scraper-go/internal/pkg/product"
	productStore "scraper-go/internal/pkg/product/store"

	"github.com/rs/zerolog/log"

	"gorm.io/gorm"
)

type store struct {
	database *gorm.DB
}

func New(
	database *gorm.DB,
) productStore.Store {
	return &store{
		database: database,
	}
}

func (s *store) CreateOne(request productStore.CreateOneRequest) (*productStore.CreateOneResponse, error) {
	cO := request.Product

	result := s.database.Create(&cO)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error creating product")
		return nil, result.Error
	}

	return &productStore.CreateOneResponse{ID: cO.ID}, nil
}

func (s *store) UpsertOne(request productStore.UpsertOneRequest) (*productStore.UpsertOneResponse, error) {
	cO := request.Product

	result := s.database.FirstOrInit(&product.Product{
		Url: request.Product.Url,
	})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error creating/updating product")
		return nil, result.Error
	}

	return &productStore.UpsertOneResponse{ID: cO.ID}, nil
}

func (s *store) FindOne(request productStore.FindOneRequest) (*productStore.FindOneResponse, error) {
	var p = product.Product{ID: request.ID}
	result := s.database.Model(product.Product{}).First(&p)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("error finding product")
		return nil, result.Error
	}

	return &productStore.FindOneResponse{Product: p}, nil
}
