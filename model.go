package rest

import (
	"net/http"
	"reflect"
)

// Serializer - could be used for JSON marshalling and unmarshalling
type Serializer interface {
	Decode() error
	Encode(v interface{}) ([]byte, error)
	UseContext(*Context)
}

// Validator - input validation
type Validator interface {
	Validate() error
	UseContext(*Context)
}

// Storage - database abstraction
type Storage interface {
	InsertOne() error
	InsertMany() error
	Remove() error
	FindOne() error
	FindMany() error
	Update() error
	Upsert() error
	UseContext(*Context)
}

// Response - holds the data to be sent to the client
type Response struct {
	Body    interface{}
	Headers map[string]string
	Status  int
}

// Context -
type Context struct {
	Action   string
	Input    interface{}
	Request  *http.Request
	Response Response
	Type     reflect.Type
}

// Model -
type Model struct {
	Name string
	Context
	Storage
	Validator
	Serializer
}

// UseStorage -
func (model *Model) UseStorage(s Storage) {
	s.UseContext(&model.Context)
	model.Storage = s
}

// UseSerializer -
func (model *Model) UseSerializer(s Serializer) {
	s.UseContext(&model.Context)
	model.Serializer = s
}

// UseValidator -
func (model *Model) UseValidator(s Validator) {
	s.UseContext(&model.Context)
	model.Validator = s
}

// Resource -
type Resource struct {
	Name       string
	Type       reflect.Type
	Headers    map[string]string
	Storage    Storage
	Validator  Validator
	Serializer Serializer
}

// New -
func (f *Resource) New(r *http.Request, action string) *Model {
	context := Context{Action: action, Request: r, Response: Response{Headers: f.Headers}, Type: f.Type}
	model := Model{}
	model.Name = f.Name
	model.Context = context
	model.UseStorage(f.Storage)
	model.UseValidator(f.Validator)
	model.UseSerializer(f.Serializer)
	return &model
}

// UseType -
func (f *Resource) UseType(t reflect.Type) *Resource {
	f.Type = t
	return f
}

// UseHeaders -
func (f *Resource) UseHeaders(headers map[string]string) *Resource {
	f.Headers = headers
	return f
}

// UseStorage -
func (f *Resource) UseStorage(s Storage) *Resource {
	f.Storage = s
	return f
}

// UseValidator -
func (f *Resource) UseValidator(v Validator) *Resource {
	f.Validator = v
	return f
}

// UseSerializer -
func (f *Resource) UseSerializer(s Serializer) *Resource {
	f.Serializer = s
	return f
}

// NewResource -
func NewResource(name string) *Resource {
	f := &Resource{Name: name}
	return f
}
