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
	url         string
	body        string
	fail_db     bool
	fail_broker bool
	nil_broker  bool
	nil_metrics bool
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
	fail_db bool
	Name    string `json:"name"`
	Age     int    `json:"age"`
}

func (tm *TestModel) Create(r *http.Request) ([]byte, error) {
	if tm.fail_db {
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
	if tm.fail_db {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func (tm *TestModel) Update(r *http.Request) ([]byte, error) {
	if tm.fail_db {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func (tm *TestModel) Delete(r *http.Request) error {
	if tm.fail_db {
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
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusCreated},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
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
	model := &TestModel{fail_db: scenario.fail_db}
	service := Service{logger: MockLogger{}, metrics: nil, broker: nil}
	if !scenario.nil_broker {
		service.broker = MockBroker{fail: scenario.fail_broker}
	}
	if !scenario.nil_metrics {
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
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: vb, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: url, body: ib, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
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
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusNoContent},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusInternalServerError},
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
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: iurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusBadRequest},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: false, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: true, nil_broker: false, nil_metrics: false}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: false, nil_broker: true, nil_metrics: false}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: false, nil_broker: false, nil_metrics: true}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: false, fail_broker: true, nil_broker: true, nil_metrics: true}), http.StatusOK},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: false, nil_broker: true, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: true, nil_broker: false, nil_metrics: true}), http.StatusInternalServerError},
		{NewTestInput(Scenario{url: vurl, fail_db: true, fail_broker: true, nil_broker: true, nil_metrics: false}), http.StatusInternalServerError},
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
