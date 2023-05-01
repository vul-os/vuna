package scraper

import (
	woocommerce "scraper-go/internal/pkg/datapoint/scraper/woocommerce"
	variationStore "scraper-go/internal/pkg/variation/store"
	productStore "scraper-go/internal/pkg/product/store"
	dpStore "scraper-go/internal/pkg/datapoint/store"

	"github.com/go-chi/chi/v5"
)

type scraper struct {
	ps productStore.Store
	vs variationStore.Store
	dps dpStore.Store
}

func New(
	ps productStore.Store,
	vs variationStore.Store,
	dps dpStore.Store,
) *scraper {
	return &scraper{
		ps: ps,
		vs: vs,
		dps: dps,
	}
}

func (a scraper) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/woocommerce", woocommerce.ScrapeOne) // POST /datapoint/scrape - scrape a single url

	return r
}
