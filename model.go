package rest

import "net/http"

// ModelFactory - has a New method that creates new model instances
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

// Identity - returns name of the model
type Identity interface {
	Name() string
}

// Serializer - could be used for JSON marshalling and unmarshalling
type Serializer interface {
	Decode() error
	Encode(v interface{}) ([]byte, error)
}

// Sanitizer - input validation
type Sanitizer interface {
	Validate() error
}

// Storage - database abstraction
type Storage interface {
	Create() (interface{}, error)
	Remove() (interface{}, error)
	Find() (interface{}, error)
	Update() (interface{}, error)
	Upsert() (interface{}, error)
}
