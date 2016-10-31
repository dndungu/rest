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
	Response() *Response
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

// Error - wraps validation errors to provide more info to the client e.g invalid fields, conflict etc
type Error struct {
	Code    int
	Message string
}

// Error - return the error message string
func (e Error) Error() string {
	return e.Message
}

// Sanitizer - input validation
type Sanitizer interface {
	Validate(mode string) *Error
}

// Storage - database abstraction
type Storage interface {
	Insert() error
	Remove() error
	FindOne() error
	FindMany() error
	Update() error
	Upsert() error
}
