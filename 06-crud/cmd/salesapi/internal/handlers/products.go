package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ardanlabs/service-training/06-crud/internal/products"
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
)

type Products struct {
	db *sqlx.DB

	http.Handler
}

func NewProducts(db *sqlx.DB) *Products {
	p := Products{db: db}

	r := chi.NewRouter()
	r.Post("/v1/products", p.Create)
	r.Get("/v1/products", p.List)
	r.Get("/v1/products/{id}", p.Get)
	p.Handler = r

	return &p
}

func (s *Products) Create(w http.ResponseWriter, r *http.Request) {
	var p products.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Printf("error: decoding product: %s", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := products.Create(s.db, &p); err != nil {
		log.Printf("error: creating product: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("error: encoding response: %s", err)
		return
	}
}

func (s *Products) List(w http.ResponseWriter, r *http.Request) {
	list, err := products.List(s.db)
	if err != nil {
		log.Printf("error: listing products: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Don't return an array (return an object with an array).
	//       Make a named response type.
	if err := json.NewEncoder(w).Encode(list); err != nil {
		log.Printf("error: encoding response: %s", err)
		return
	}
}

func (s *Products) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	p, err := products.Get(s.db, id)
	if err != nil {
		log.Printf("error: getting product: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Printf("error: encoding response: %s", err)
		return
	}
}