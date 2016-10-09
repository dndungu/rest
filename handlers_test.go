package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Scenario struct {
	url        string
	body       string
	failDB     bool
	failBroker bool
	nilBroker  bool
	nilMetrics bool
}

type MockBroker struct {
	fail bool
}

func (mb MockBroker) Publish(event string, v interface{}) error {
	if mb.fail {
		return errors.New("The broker failed on purpose")
	}
	return nil
}

type MockLogger struct{}

func (ml MockLogger) Info(v interface{})    {}
func (ml MockLogger) Warning(v interface{}) {}
func (ml MockLogger) Error(v interface{})   {}
func (ml MockLogger) Fatal(v interface{})   {}

type MockMetrics struct {
	fail bool
}

func (mm MockMetrics) Incr(stat string, count int64)   {}
func (mm MockMetrics) Timing(stat string, delta int64) {}
func (mm MockMetrics) NewTimer(stat string) func() {
	return func() {}
}

type TestModel struct {
	failDB bool
	Name   string `json:"name"`
	Age    int    `json:"age"`
}

func (tm *TestModel) Create(r *http.Request) ([]byte, error) {
	if tm.failDB {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func (tm *TestModel) Validate(r *http.Request) error {
	if strings.Compare(r.Method, "DELETE") == 0 || strings.Compare(r.Method, "GET") == 0 {
		if strings.Compare(r.URL.Path, "/test/1") != 0 {
			return errors.New("Invalid URL parameter")
		}
		return nil
	}
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(tm)
}

func (tm *TestModel) Find(r *http.Request) ([]byte, error) {
	if tm.failDB {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func (tm *TestModel) Update(r *http.Request) ([]byte, error) {
	if tm.failDB {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func (tm *TestModel) Delete(r *http.Request) error {
	if tm.failDB {
		return errors.New("Database failed on purpose")
	}
	return nil
}

type TestInput struct {
	url     string
	body    string
	model   *TestModel
	service Service
}

func TestInsert(t *testing.T) {
	vb := `{"name": "Otieno Kamau", "age": 21}`
	ib := "name=Bad Name"
	url := "http://foo.bar/test"
	tests := []struct {
		input    TestInput
		expected int
	}{
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
	}
	for _, test := range tests {
		h := test.input.service.Insert("test", test.input.model)
		w := httptest.NewRecorder()
		r := NewTestRequest("POST", test.input.url, test.input.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.input.body)
		}
	}
}

func NewTestInput(scenario Scenario) TestInput {
	model := &TestModel{failDB: scenario.failDB}
	service := Service{logger: MockLogger{}, metrics: nil, broker: nil}
	if !scenario.nilBroker {
		service.broker = MockBroker{fail: scenario.failBroker}
	}
	if !scenario.nilMetrics {
		service.metrics = MockMetrics{}
	}
	return TestInput{scenario.url, scenario.body, model, service}
}

func NewTestRequest(verb, url, input string) *http.Request {
	return httptest.NewRequest(verb, url, bytes.NewBufferString(input))
}

func TestUpdate(t *testing.T) {
	vb := `{"name": "Otieno Kamau", "age": 21}`
	ib := "name=Bad Name"
	url := "http://foo.bar/test/1"
	tests := []struct {
		input    TestInput
		expected int
	}{
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
	}
	for _, test := range tests {
		h := test.input.service.Update("test", test.input.model)
		w := httptest.NewRecorder()
		r := NewTestRequest("PUT", test.input.url, test.input.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.input.body)
		}
	}
}

func TestDelete(t *testing.T) {
	vurl := "http://foo.bar/test/1"
	iurl := "http://foo.bar/test"
	tests := []struct {
		input    TestInput
		expected int
	}{
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusInternalServerError},
	}
	for _, test := range tests {
		h := test.input.service.Delete("test", test.input.model)
		w := httptest.NewRecorder()
		r := NewTestRequest("DELETE", test.input.url, test.input.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.input.url)
		}
	}
}

func TestFind(t *testing.T) {
	vurl := "http://foo.bar/test/1"
	iurl := "http://foo.bar/test/bad-id-format"
	tests := []struct {
		input    TestInput
		expected int
	}{
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}), http.StatusInternalServerError},
	}
	for _, test := range tests {
		h := test.input.service.Find("test", test.input.model)
		w := httptest.NewRecorder()
		r := NewTestRequest("GET", test.input.url, test.input.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Error(test.input)
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.input.url)
		}
	}
}
