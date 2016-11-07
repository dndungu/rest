package rest

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// JSON -
type JSON struct {
	*Context
}

// UseContext -
func (j *JSON) UseContext(c *Context) {
	j.Context = c
}

// Decode -
func (j *JSON) Decode() error {
	if j.Context.Request.ContentLength == 0 {
		return nil
	}
	v := reflect.New(j.Context.Type).Interface()
	decoder := json.NewDecoder(j.Context.Request.Body)
	err := decoder.Decode(&v)
	j.Context.Input = v
	if err != nil {
		j.Context.Response.Status = http.StatusBadRequest
		j.Context.Response.Body = err.Error()
	}
	return err
}

// Encode -
func (j *JSON) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
