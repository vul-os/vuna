package adaptor

import (
	"net/http"

	site "scraper-go/internal/pkg/site"
	siteStore "scraper-go/internal/pkg/site/store"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type api struct {
	store siteStore.Store
}

func New(
	ps siteStore.Store,
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
	resp, err := a.store.FindOne(siteStore.FindOneRequest{
		ID: u,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.Site)
	return
}

func (a api) CreateOne(w http.ResponseWriter, r *http.Request) {
	var s site.Site

	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.store.CreateOne(siteStore.CreateOneRequest{
		Site: s,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.ID)
	return
}
