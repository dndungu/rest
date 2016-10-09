package rest

import "net/http"

// Model defines the interface models are expected to expose
type Model interface {
	Create(r *http.Request) ([]byte, error)
	Delete(r *http.Request) error
	Find(r *http.Request) ([]byte, error)
	Update(r *http.Request) ([]byte, error)
	Validate(r *http.Request) error
}
