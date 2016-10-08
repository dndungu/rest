package rest

import (
	"testing"
)

func TestInternalServerErrorResponse(t *testing.T) {
	status, _ := InternalServerErrorResponse()
	expected := 500
	if status != expected {
		t.Errorf("Error, expected %d got %d", expected, status)
	}
}

func TestCreatedResponse(t *testing.T) {
	status, _ := CreatedResponse()
	expected := 201
	if status != expected {
		t.Errorf("Error, expected %d got %d", expected, status)
	}
}

func TestBadRequestResponse(t *testing.T) {
	status, _ := BadRequestResponse()
	expected := 400
	if status != expected {
		t.Errorf("Error, expected %d got %d", expected, status)
	}
}

func TestNoContentResponse(t *testing.T) {
	status, _ := NoContentResponse()
	expected := 204
	if status != expected {
		t.Errorf("Error, expected %d got %d", expected, status)
	}
}

func TestNotFoundResponse(t *testing.T) {
	status, _ := NotFoundResponse()
	expected := 404
	if status != expected {
		t.Errorf("Error, expected %d got %d", expected, status)
	}
}
