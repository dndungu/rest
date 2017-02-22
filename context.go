package rest

import "net/http"

const (
	// REQUEST - the key for http.Request object in the context
	REQUEST = "request"
	// RESPONSE - the key for Response object in the context
	RESPONSE = "response"
	// ACTION - the data operation being carried out in the transaction
	ACTION = "action"
	// DATATYPE - the reflect.Type of the resource in the transaction
	DATATYPE = "type"
	// REQUESTBODY - the data sent from the client
	REQUESTBODY = "requestBody"
)

// Context -
type Context struct {
	data map[string]interface{}
}

// NewContext -
func NewContext() Context {
	c := Context{}
	c.data = make(map[string]interface{})
	return c
}

// Get -
func (c *Context) Get(key string) (i interface{}) {
	return c.data[key]
}

// Set -
func (c *Context) Set(key string, value interface{}) *Context {
	c.data[key] = value
	return c
}

// GetRequest -
func (c *Context) GetRequest() (r *http.Request) {
	return c.data[REQUEST].(*http.Request)
}

// GetResponse -
func (c *Context) GetResponse() (r Response) {
	return c.data[RESPONSE].(Response)
}

func (c *Context) SetResponse(r Response) {
	c.data[RESPONSE] = r
}

// SetResponseBody -
func (c *Context) SetResponseBody(b interface{}) {
	response := c.GetResponse()
	response.Body = b
	c.SetResponse(response)
}

// SetResponseHeaders -
func (c *Context) SetResponseHeaders(h map[string][]string) {
	response := c.GetResponse()
	response.Headers = h
	c.SetResponse(response)
}

// SetResponseStatus -
func (c *Context) SetResponseStatus(s int) {
	response := c.GetResponse()
	response.Status = s
	c.SetResponse(response)
}
