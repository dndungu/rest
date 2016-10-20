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

type FakeModel struct {
	request *http.Request
	FakeStorage
	FakeIdentity
	Fields FakeModelFields
	items  []FakeModelFields
}

type FakeIdentity struct {
	name string
}

func (fi *FakeIdentity) Name() string {
	return fi.name
}

func (fm *FakeModel) Validate() error {
	method := fm.request.Method
	if strings.Compare(method, "DELETE") == 0 || strings.Compare(method, "GET") == 0 {
		path := fm.request.URL.Path
		if strings.Compare(path, "/test/1") != 0 {
			return errors.New("Invalid URL parameter")
		}
		return nil
	}
	if fm.items[0].Name == `Otieno Kamau` && fm.items[0].Age == 21 {
		return nil
	}
	return errors.New("The data is invalid")
}

func (fm *FakeModel) Decode() error {
	fm.items = []FakeModelFields{{}}
	decoder := json.NewDecoder(fm.request.Body)
	return decoder.Decode(&fm.items[0])
}

func (fm *FakeModel) Encode(v interface{}) ([]byte, error) {
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

func (fs *FakeStorage) Find() (interface{}, error) {
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
	service := Service{logger: MockLogger{}, metrics: nil, broker: nil}
	if !scenario.nilBroker {
		service.broker = MockBroker{fail: scenario.failBroker}
	}
	if !scenario.nilMetrics {
		service.metrics = MockMetrics{}
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
	return &FakeModel{request: r, FakeStorage: FakeStorage{fail: fmf.fail}}
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
	iurl := "http://foo.bar/test"
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

func TestFind(t *testing.T) {
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
		h := service.Find(mf)
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
