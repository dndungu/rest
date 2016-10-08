package rest

import (
	"net/http"
)

func InternalServerErrorResponse() (int, []byte) {
	i := http.StatusInternalServerError
	return i, []byte(http.StatusText(i))
}

func CreatedResponse() (int, []byte) {
	i := http.StatusCreated
	return i, []byte(http.StatusText(i))
}

func BadRequestResponse() (int, []byte) {
	i := http.StatusBadRequest
	return i, []byte(http.StatusText(i))
}

func NoContentResponse() (int, []byte) {
	i := http.StatusNoContent
	return i, []byte(http.StatusText(i))
}

func NotFoundResponse() (int, []byte) {
	i := http.StatusNotFound
	return i, []byte(http.StatusText(i))
}
