package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type scraper struct{}

// Routes creates a REST router for the todos resource
func (s scraper) Routes() chi.Router {
	r := chi.NewRouter()
	// r.Use() // some middleware..

	// r.Get("/", s.List)   
	// r.Post("/", s.List)   

	r.Route("/store/{id}", func(r chi.Router) {
		r.Get("/", rs.Get)       // GET /store/{id} - read a single todo by :id
	})

	r.Route("/product/{id}", func(r chi.Router) {
		r.Get("/datapoint", rs.Get)       // GET /todos/{id} - read a single todo by :id
		r.Get("/otherdata", rs.Update)    // PUT /todos/{id} - update a single todo by :id
	})

	return r
}

func (rs scraper) List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todos list of stuff.."))
}

func (rs scraper) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todos create"))
}
