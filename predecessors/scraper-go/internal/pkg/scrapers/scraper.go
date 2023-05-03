package scraper

import (
	dpStore "scraper-go/internal/pkg/datapoint/store"
	woocommerce "scraper-go/internal/pkg/product/scraper/woocommerce"
	productStore "scraper-go/internal/pkg/product/store"
	variationStore "scraper-go/internal/pkg/variation/store"

	"github.com/go-chi/chi/v5"
)

type scraper struct {
	ps  productStore.Store
	vs  variationStore.Store
	dps dpStore.Store
}

func New(
	ps productStore.Store,
	vs variationStore.Store,
	dps dpStore.Store,
) *scraper {
	return &scraper{
		ps:  ps,
		vs:  vs,
		dps: dps,
	}
}

func (a scraper) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/scraper", woocommerce.ScrapeOne) // POST /datapoint/scrape - scrape a single url

	return r
}
