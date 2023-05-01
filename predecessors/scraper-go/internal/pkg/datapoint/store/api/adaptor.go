package adaptor

import (
	"net/http"

	dp "scraper-go/internal/pkg/datapoint"
	dpStore "scraper-go/internal/pkg/datapoint/store"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type api struct {
	store dpStore.Store
}

func New(
	dps dpStore.Store,
) *api {
	return &api{
		store: dps,
	}
}

// Routes creates a REST router for the products resource
func (a api) Routes() chi.Router {
	r := chi.NewRouter()

	// r.Get("/", a.List)    // GET /products - read a list of todos
	r.Post("/", a.CreateOne) // Post /products - create a product

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", a.FindOne) // GET /products/{id} - read a single products by :id
		// r.Delete("/", a.Delete) // DELETE /products/{id} - delete a single products by :id
	})

	return r
}

func (a api) FindOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := a.store.FindOne(dpStore.FindOneRequest{
		ID: id,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.Datapoint)
}

func (a api) CreateOne(w http.ResponseWriter, r *http.Request) {
	var dp dp.DataPoint

	err := json.NewDecoder(r.Body).Decode(&dp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.store.CreateOne(dpStore.CreateOneRequest{
		Datapoint: dp,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.ID)
}
