package rest

import (
	"net/http"
)

// Response - holds the data to be sent to the client
type Response struct {
	Body    interface{}
	Headers map[string]string
	Status  int
}

// InternalServerErrorResponse returns 500, Internal Server Error
func InternalServerErrorResponse() (int, []byte) {
	i := http.StatusInternalServerError
	return i, []byte(http.StatusText(i))
}

// CreatedResponse returns 201, Created
func CreatedResponse() (int, []byte) {
	i := http.StatusCreated
	return i, []byte(http.StatusText(i))
}

// BadRequestResponse returns 400, Bad Request
func BadRequestResponse() (int, []byte) {
	i := http.StatusBadRequest
	return i, []byte(http.StatusText(i))
}

// NoContentResponse returns 204, No Content
func NoContentResponse() (int, []byte) {
	i := http.StatusNoContent
	return i, []byte(http.StatusText(i))
}

// NotFoundResponse returns 404, Not found
func NotFoundResponse() (int, []byte) {
	i := http.StatusNotFound
	return i, []byte(http.StatusText(i))
}
