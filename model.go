package rest

import "net/http"

// ModelFactory
type ModelFactory interface {
	New(r *http.Request) Model
}

// Model defines the interface models are expected to expose
type Model interface {
	Identity
	Sanitizer
	Serializer
	Storage
}

// Identity
type Identity interface {
	Name() string
}

// Serializer
type Serializer interface {
	Decode() error
	Encode(v interface{}) ([]byte, error)
}

// Sanitizer
type Sanitizer interface {
	Validate() error
}

// Storage
type Storage interface {
	Create() (interface{}, error)
	Remove() (interface{}, error)
	Find() (interface{}, error)
	Update() (interface{}, error)
	Upsert() (interface{}, error)
}
