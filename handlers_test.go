package rest

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type FakeScenario struct {
	url          string
	body         string
	failDatabase bool
	failBroker   bool
	failMetrics  bool
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

func (ml MockLogger) Info(v interface{}) {
}
func (ml MockLogger) Warning(v interface{}) {
}
func (ml MockLogger) Error(v interface{}) {
}
func (ml MockLogger) Fatal(v interface{}) {
}

type MockMetrics struct {
	fail bool
}

func (mm MockMetrics) Incr(stat string, count int64) error {
	if mm.fail {
		return errors.New("The metrics client failed on purpose")
	}
	return nil
}

func (mm MockMetrics) Timing(stat string, delta int64) error {
	if mm.fail {
		return errors.New("The metrics client failed on purpose")
	}
	return nil
}

func (mm MockMetrics) NewTimer(stat string) func() {
	return func() {}
}

type FakeFields struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type FakeValidator struct {
	*Context
}

func (v *FakeValidator) UseContext(c *Context) {
	v.Context = c
}

func (v *FakeValidator) Validate() error {
	var msg string
	if v.Request.URL.Path == "/test/bad-id-format" {
		msg = "Invalid URL parameter"
		v.Response.Status = 400
		v.Response.Body = msg
		return errors.New(msg)
	}
	if v.Action == "insert_one" || v.Action == "update" || v.Action == "upsert" {
		input := v.Input.(*FakeFields)
		if input.Name == `Otieno Kamau` && input.Age == 21 {
			return nil
		}
		msg = "The data is invalid"
		v.Response.Status = 400
		v.Response.Body = msg
		return errors.New(msg)
	}
	return nil
}

type FakeMessage struct {
	message string
}

type FakeStorage struct {
	fail    bool
	Context *Context
}

func (fs *FakeStorage) UseContext(c *Context) {
	fs.Context = c
}

func (fs *FakeStorage) InsertOne() error {
	fs.Context.Response.Body = fs.Context.Input
	return fs.FakeAction(http.StatusCreated, http.StatusInternalServerError)
}

func (fs *FakeStorage) InsertMany() error {
	fs.Context.Response.Body = fs.Context.Input
	return fs.FakeAction(http.StatusCreated, http.StatusInternalServerError)
}

func (fs *FakeStorage) FindOne() error {
	return fs.FakeAction(http.StatusOK, http.StatusInternalServerError)
}

func (fs *FakeStorage) FindMany() error {
	return fs.FakeAction(http.StatusOK, http.StatusInternalServerError)
}

func (fs *FakeStorage) Update() error {
	fs.Context.Response.Body = fs.Context.Input
	return fs.FakeAction(http.StatusNoContent, http.StatusInternalServerError)
}

func (fs *FakeStorage) Remove() error {
	return fs.FakeAction(http.StatusNoContent, http.StatusInternalServerError)
}

func (fs *FakeStorage) Upsert() error {
	fs.Context.Response.Body = fs.Context.Input
	return fs.FakeAction(http.StatusOK, http.StatusInternalServerError)
}

func (fs *FakeStorage) FakeAction(good, bad int) error {
	if fs.fail {
		fs.Context.Response.Status = bad
		return errors.New("Database failed on purpose")
	}
	fs.Context.Response.Status = good
	return nil

}

func NewFakeService(scenario FakeScenario) *Service {
	service := NewService()
	service.UseLogger(&MockLogger{})
	service.UseBroker(&MockBroker{fail: scenario.failBroker})
	service.UseMetrics(&MockMetrics{fail: scenario.failMetrics})
	return service
}

func NewTestRequest(verb, url, input string) *http.Request {
	var body io.Reader
	if input != "" {
		body = bytes.NewBufferString(input)
	}
	return httptest.NewRequest(verb, url, body)
}

func NewFakeFactory(s FakeScenario) *ModelFactory {
	headers := map[string]string{"Content-Type": "application/json"}
	f := NewFactory("tester").
		SetDefaultHeaders(headers).
		UseType(reflect.TypeOf(FakeFields{})).
		UseStorage(&FakeStorage{fail: s.failDatabase}).
		UseValidator(&FakeValidator{}).
		UseSerializer(&JSON{})
	return f
}

func TestInsertOne(t *testing.T) {
	validBody := `{"name": "Otieno Kamau", "age": 21}`
	invalidJSON := "bad body"
	invalidBody := `{"name": "Otieno Kamau", "age": 12}`
	url := "http://foo.bar/test"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusCreated},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},

		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},

		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},
	}
	for _, test := range tests {
		service := NewFakeService(test.scenario)
		f := NewFakeFactory(test.scenario)
		h := service.InsertOne(f)
		w := httptest.NewRecorder()
		r := NewTestRequest("POST", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.scenario.body)
		}
	}
}

func TestInsertMany(t *testing.T) {
	validBody := `[{"name": "Otieno Kamau", "age": 21}, {"name": "Bernie Burst", "age": 81}]`
	url := "http://foo.bar/test"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusCreated},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		service := NewFakeService(test.scenario)
		f := NewFakeFactory(test.scenario)
		h := service.InsertMany(f)
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
	validBody := `{"name": "Otieno Kamau", "age": 21}`
	invalidJSON := "name=Bad Name"
	invalidBody := `{"name": "Otieno Kamau", "age": 17}`
	url := "http://foo.bar/test/1"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},

		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},

		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},
	}
	for _, test := range tests {
		f := NewFakeFactory(test.scenario)
		service := NewFakeService(test.scenario)
		h := service.Update(f)
		w := httptest.NewRecorder()
		r := NewTestRequest("PUT", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d for request body: %s", test.expected, actual, test.scenario.body)
		}
	}
}

func TestUpsert(t *testing.T) {
	validBody := `{"name": "Otieno Kamau", "age": 21}`
	invalidJSON := "name=Bad Name"
	invalidBody := `{"name": "Otieno Kamau", "age": 17}`
	url := "http://foo.bar/test/1"
	tests := []struct {
		scenario FakeScenario
		expected int
	}{
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusOK},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusOK},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: url, body: validBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},

		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidJSON, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},

		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: url, body: invalidBody, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},
	}
	for _, test := range tests {
		f := NewFakeFactory(test.scenario)
		service := NewFakeService(test.scenario)
		h := service.Upsert(f)
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
		{FakeScenario{url: iurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusNoContent},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		f := NewFakeFactory(test.scenario)
		service := NewFakeService(test.scenario)
		h := service.Remove(f)
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
		{FakeScenario{url: iurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusBadRequest},
		{FakeScenario{url: iurl, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusBadRequest},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		f := NewFakeFactory(test.scenario)
		service := NewFakeService(test.scenario)
		h := service.FindOne(f)
		w := httptest.NewRecorder()
		r := NewTestRequest("GET", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
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
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: false}, http.StatusOK},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: false, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: false, failBroker: true, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: false, failMetrics: true}, http.StatusInternalServerError},
		{FakeScenario{url: vurl, failDatabase: true, failBroker: true, failMetrics: false}, http.StatusInternalServerError},
	}
	for _, test := range tests {
		f := NewFakeFactory(test.scenario)
		service := NewFakeService(test.scenario)
		h := service.FindMany(f)
		w := httptest.NewRecorder()
		r := NewTestRequest("GET", test.scenario.url, test.scenario.body)
		h(w, r)
		actual := w.Code
		if actual != test.expected {
			t.Errorf("Error, expected %d, got %d using URL: %s", test.expected, actual, test.scenario.url)
		}
	}
}
