package adaptor

import (
	"net/http"

	varitaion "scraper-go/internal/pkg/variation"
	variationStore "scraper-go/internal/pkg/variation/store"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type api struct {
	store variationStore.Store
}

func New(
	ps variationStore.Store,
) *api {
	return &api{
		store: ps,
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
	u, err := uuid.Parse(id)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	resp, err := a.store.FindOne(variationStore.FindOneRequest{
		ID: u,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.Variation)
}

func (a api) CreateOne(w http.ResponseWriter, r *http.Request) {
	var v varitaion.Variation

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.store.CreateOne(variationStore.CreateOneRequest{
		Variation: v,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.ID)
}
