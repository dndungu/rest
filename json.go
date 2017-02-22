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
func (j *JSON) Decode() (err error) {
	r := j.Context.GetRequest()
	if r.Method != "POST" && r.Method != "PUT" && r.Method != "PATCH" {
		return nil
	}
	t := j.Context.Get(DATATYPE).(reflect.Type)
	v := reflect.New(t).Interface()
	if j.Context.Get(ACTION) == INSERTMANY {
		v = reflect.New(reflect.SliceOf(t)).Interface()
	}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&v)
	j.Context.Set(REQUESTBODY, v)
	if err != nil {
		j.Context.SetResponseStatus(http.StatusBadRequest)
		j.Context.SetResponseBody(err.Error())
	}
	return err
}

// Encode -
func (j *JSON) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
