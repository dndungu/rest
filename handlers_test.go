package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type FakeScenario struct {
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

type FakeModelFields struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type FakeIdentity struct {
	name string
}

func (fi *FakeIdentity) Name() string {
	return fi.name
}

type FakeModel struct {
	data    FakeModelFields
	request *http.Request
	FakeStorage
	FakeIdentity
	FakeSerializer
}

func (fm *FakeModel) Validate() *ValidationError {
	method := fm.request.Method
	if fm.request.URL.Path == "/test/bad-id-format" {
		return &ValidationError{400, "Invalid URL parameter"}
	}
	if method == "POST" || method == "PUT" {
		if fm.data.Name == `Otieno Kamau` && fm.data.Age == 21 {
			return nil
		}
		return &ValidationError{400, "The data is invalid"}
	}
	return nil
}

type FakeSerializer struct {
	request *http.Request
	data    *FakeModelFields
}

func (fs *FakeSerializer) Decode() error {
	decoder := json.NewDecoder(fs.request.Body)
	return decoder.Decode(fs.data)
}

func (fs *FakeSerializer) Encode(v interface{}) ([]byte, error) {
	return nil, nil
}

type FakeMessage struct {
	message string
}

type FakeStorage struct {
	fail bool
}

func (fs *FakeStorage) Create() (interface{}, error) {
	return fs.FakeAction()
}

func (fs *FakeStorage) FindOne() (interface{}, error) {
	return fs.FakeAction()
}

func (fs *FakeStorage) FindMany() (interface{}, error) {
	return fs.FakeAction()
}

func (fs *FakeStorage) Update() (interface{}, error) {
	return fs.FakeAction()
}

func (fs *FakeStorage) Remove() (interface{}, error) {
	return fs.FakeAction()
}

func (fs *FakeStorage) Upsert() (interface{}, error) {
	return fs.FakeAction()
}

func (fs *FakeStorage) FakeAction() (interface{}, error) {
	if fs.fail {
		return nil, errors.New("Database failed on purpose")
	}
	return nil, nil
}

func NewFakeService(scenario FakeScenario) *Service {
	service := Service{Logger: MockLogger{}, Metrics: nil, Broker: nil}
	if !scenario.nilBroker {
		service.Broker = MockBroker{fail: scenario.failBroker}
	}
	if !scenario.nilMetrics {
		service.Metrics = MockMetrics{}
	}
	return &service
}

func NewTestRequest(verb, url, input string) *http.Request {
	return httptest.NewRequest(verb, url, bytes.NewBufferString(input))
}

type FakeModelFactory struct {
	fail bool
}

func (fmf FakeModelFactory) New(r *http.Request) Model {
	model := FakeModel{FakeModelFields{}, r, FakeStorage{fmf.fail}, FakeIdentity{"test"}, FakeSerializer{request: r}}
	model.FakeSerializer.data = &model.data
	return &model
}

func TestInsert(t *testing.T) {
	vb := `{"name": "Otieno Kamau", "age": 21}`
	bb := "bad body"
	ib := `{"name": "Otieno Kamau", "age": 12}`
	url := "http://foo.bar/test"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusCreated},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusCreated},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusCreated},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusCreated},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusInternalServerError},

		{FakeScenario{url: url, body: bb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},

		{FakeScenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
	}
	for _, test := range tests {
		service := NewFakeService(test.scenario)
		mf := FakeModelFactory{fail: test.scenario.failDB}
		h := service.Insert(mf)
		w := httptest.NewRecorder()
		r := NewTestRequest("POST", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.scenario.body)
		}
	}
}

func TestUpdate(t *testing.T) {
	vb := `{"name": "Otieno Kamau", "age": 21}`
	bb := "name=Bad Name"
	ib := `{"name": "Otieno Kamau", "age": 17}`
	url := "http://foo.bar/test/1"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusNoContent},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusNoContent},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: vb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusInternalServerError},

		{FakeScenario{url: url, body: bb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: bb, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},

		{FakeScenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: ib, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
	}
	for _, test := range tests {
		mf := FakeModelFactory{fail: test.scenario.failDB}
		service := NewFakeService(test.scenario)
		h := service.Update(mf)
		w := httptest.NewRecorder()
		r := NewTestRequest("PUT", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.scenario.body)
		}
	}
}

func TestRemove(t *testing.T) {
	vurl := "http://foo.bar/test/1"
	iurl := "http://foo.bar/test/bad-id-format"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusNoContent},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusNoContent},
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		mf := FakeModelFactory{fail: test.scenario.failDB}
		service := NewFakeService(test.scenario)
		h := service.Remove(mf)
		w := httptest.NewRecorder()
		r := NewTestRequest("DELETE", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.scenario.url)
		}
	}
}

func TestFindOne(t *testing.T) {
	vurl := "http://foo.bar/test/1"
	iurl := "http://foo.bar/test/bad-id-format"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		mf := FakeModelFactory{fail: test.scenario.failDB}
		service := NewFakeService(test.scenario)
		h := service.FindOne(mf)
		w := httptest.NewRecorder()
		r := NewTestRequest("GET", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Error(test.scenario)
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.scenario.url)
		}
	}
}

func TestFindMany(t *testing.T) {
	vurl := "http://foo.bar/test"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: true, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: true, nilMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: false, failBroker: false, nilBroker: false, nilMetrics: true}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: false, failBroker: true, nilBroker: true, nilMetrics: true}, http.StatusOK},
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: true, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: false, nilMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDB: true, failBroker: true, nilBroker: true, nilMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		mf := FakeModelFactory{fail: test.scenario.failDB}
		service := NewFakeService(test.scenario)
		h := service.FindMany(mf)
		w := httptest.NewRecorder()
		r := NewTestRequest("GET", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Error(test.scenario)
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.scenario.url)
		}
	}
}

func TestFindBadMode(t *testing.T) {
	vurl := "http://foo.bar/test"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: vurl, failDB: true, failBroker: false, nilBroker: false, nilMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		mf := FakeModelFactory{fail: test.scenario.failDB}
		service := NewFakeService(test.scenario)
		h := service.find(mf, "abcd")
		w := httptest.NewRecorder()
		r := NewTestRequest("GET", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Error(test.scenario)
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.scenario.url)
		}
	}
}
