package rest

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	fail_db         bool
	fail_validation bool
	Name            string `json:"name"`
	Age             int    `json:"age"`
}

func (tm *TestModel) Create() (interface{}, error) {
	if tm.fail_db {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func (tm *TestModel) Validate() error {
	if tm.fail_validation {
		return errors.New("Validation failing on purpose")
	}
	return nil
}

func (tm *TestModel) Find(r *http.Request) ([]interface{}, error) {
	return nil, nil
}

type TestInput struct {
	body    string
	model   *TestModel
	service Service
}

func TestInsertOne(t *testing.T) {
	valid_body := `{"name": "Otieno Kamau", "age": 21}`
	invalid_body := "name=Bad Name"
	tests := []struct {
		input    TestInput
		expected int
	}{
		{TestInput{valid_body, &TestModel{fail_db: false, fail_validation: false}, Service{nil, MockLogger{}, MockMetrics{}}}, http.StatusCreated},
		{TestInput{valid_body, &TestModel{fail_db: false, fail_validation: false}, Service{MockBroker{false}, MockLogger{}, nil}}, http.StatusCreated},
		{TestInput{valid_body, &TestModel{fail_db: false, fail_validation: false}, Service{nil, MockLogger{}, nil}}, http.StatusCreated},
		{NewTestInput(invalid_body, false, false, false), http.StatusBadRequest},
		{NewTestInput(invalid_body, true, false, false), http.StatusBadRequest},
		{NewTestInput(invalid_body, false, true, false), http.StatusBadRequest},
		{NewTestInput(invalid_body, false, false, true), http.StatusBadRequest},
		{NewTestInput(invalid_body, true, true, true), http.StatusBadRequest},
		{NewTestInput(invalid_body, false, true, true), http.StatusBadRequest},
		{NewTestInput(invalid_body, true, false, true), http.StatusBadRequest},
		{NewTestInput(invalid_body, true, true, false), http.StatusBadRequest},
		{NewTestInput(valid_body, false, false, false), http.StatusCreated},
		{NewTestInput(valid_body, true, false, false), http.StatusInternalServerError},
		{NewTestInput(valid_body, false, true, false), http.StatusBadRequest},
		{NewTestInput(valid_body, false, false, true), http.StatusInternalServerError},
		{NewTestInput(valid_body, true, true, true), http.StatusBadRequest},
		{NewTestInput(valid_body, false, true, false), http.StatusBadRequest},
		{NewTestInput(valid_body, true, false, true), http.StatusInternalServerError},
		{NewTestInput(valid_body, true, true, false), http.StatusBadRequest},
	}
	for _, test := range tests {
		h := test.input.service.InsertOne("test", test.input.model)
		w := httptest.NewRecorder()
		r := NewTestRequest("POST", "http://foo.bar/test", test.input.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.input.body)
		}
	}
}

func NewTestInput(body string, fail_db, fail_validation, fail_broker bool) TestInput {
	return TestInput{body, &TestModel{fail_db: fail_db, fail_validation: fail_validation}, NewTestService(fail_broker)}
}

func NewTestService(fail bool) Service {
	return Service{MockBroker{fail}, MockLogger{}, MockMetrics{}}
}

func NewTestRequest(verb, url, input string) *http.Request {
	return httptest.NewRequest(verb, url, bytes.NewBufferString(input))
}
