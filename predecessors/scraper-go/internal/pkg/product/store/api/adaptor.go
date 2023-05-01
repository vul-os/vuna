package adaptor

import (
	"net/http"

	product "scraper-go/internal/pkg/product"
	productStore "scraper-go/internal/pkg/product/store"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type api struct {
	store productStore.Store
}

func New(
	ps productStore.Store,
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
	resp, err := a.store.FindOne(productStore.FindOneRequest{
		ID: id,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.Product)
	return
}

func (a api) CreateOne(w http.ResponseWriter, r *http.Request) {
	var p product.Product

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.store.CreateOne(productStore.CreateOneRequest{
		Product: p,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, resp.ID)
	return
}
