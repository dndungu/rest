package rest

import "net/http"

const (
	REQUEST     = "request"
	RESPONSE    = "response"
	ACTION      = "action"
	DATATYPE    = "type"
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

// Set
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

func (c *Context) SetResponseBody(b interface{}) {
	response := c.GetResponse()
	response.Body = b
	c.SetResponse(response)
}

func (c *Context) SetResponseHeaders(h map[string][]string) {
	response := c.GetResponse()
	response.Headers = h
	c.SetResponse(response)
}

func (c *Context) SetResponseStatus(s int) {
	response := c.GetResponse()
	response.Status = s
	c.SetResponse(response)
}
