package rest

import (
	"errors"
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
	Headers map[string][]string
	Status  int
}

// Model -
type Model struct {
	Name string
	Context
	Storage
	Validator
	Serializer
}

const (
	INSERTONE  = "insertOne"
	INSERTMANY = "insertMany"
	UPDATE     = "update"
	UPSERT     = "upsert"
	FINDONE    = "findOne"
	FINDMANY   = "findMany"
	REMOVE     = "remove"
)

// Execute
func (model *Model) Execute(action string) error {
	switch {
	case action == INSERTONE:
		return model.InsertOne()
	case action == INSERTMANY:
		return model.InsertMany()
	case action == UPDATE:
		return model.Update()
	case action == UPSERT:
		return model.Upsert()
	case action == FINDONE:
		return model.FindOne()
	case action == FINDMANY:
		return model.FindMany()
	case action == REMOVE:
		return model.Remove()
	}
	return errors.New("Action must be one of [insertOne, insertMany, update, upsert, findOne, findMany, remove]")
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
	Headers    map[string][]string
	Storage    Storage
	Validator  Validator
	Serializer Serializer
}

// NewModel -
func (r *Resource) NewModel(req *http.Request, action string) *Model {
	model := Model{}
	model.Name = r.Name
	model.Context = NewContext()
	model.Context.Set("action", action)
	model.Context.Set("request", req)
	model.Context.Set("response", Response{Headers: r.Headers})
	model.Context.Set("type", r.Type)
	model.UseStorage(r.Storage)
	model.UseValidator(r.Validator)
	model.UseSerializer(r.Serializer)
	return &model
}

// UseType -
func (r *Resource) UseType(t reflect.Type) *Resource {
	r.Type = t
	return r
}

// UseHeaders -
func (r *Resource) UseHeaders(headers map[string][]string) *Resource {
	r.Headers = headers
	return r
}

// UseStorage -
func (r *Resource) UseStorage(s Storage) *Resource {
	r.Storage = s
	return r
}

// UseValidator -
func (r *Resource) UseValidator(v Validator) *Resource {
	r.Validator = v
	return r
}

// UseSerializer -
func (r *Resource) UseSerializer(s Serializer) *Resource {
	r.Serializer = s
	return r
}

// NewResource -
func NewResource(name string) *Resource {
	r := &Resource{Name: name}
	return r
}
