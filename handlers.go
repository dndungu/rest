package rest

import (
	"gopkg.in/zatiti/router.v1"
	"net/http"
)

// write creates a http handler for creating or updating a document depending on the mode provided
func (s *Service) persist(modelFactory ModelFactory, mode string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// Model is request scoped
		model := modelFactory.New(r)
		// Event is the name used to track metrics
		event := model.Name() + "_" + mode
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Send response back to client
		write := func(status int, body []byte) {
			w.WriteHeader(status)
			w.Write(body)
		}
		// Instanciate a value of the model being created from the request body
		err := model.Decode()
		if err != nil {
			status, body = BadRequestResponse()
			write(status, body)
			s.Logger.Error(err)
			return
		}
		// Validate user input
		verr := model.Validate()
		if verr != nil {
			status, body = verr.Code, []byte(verr.Message)
			write(status, body)
			s.Logger.Error(verr.Message)
			return
		}
		// Call the relevant model action
		switch {
		case mode == "insert":
			err = model.Create()
		case mode == "update":
			err = model.Update()
		case mode == "upsert":
			err = model.Upsert()
		}
		// Handle failed database operation
		if err != nil {
			status, body = InternalServerErrorResponse()
			write(status, body)
			s.Logger.Error(err)
			return
		}
		response := model.Response()
		// If event broker is defined use it
		if s.Broker != nil {
			err = s.Broker.Publish(event, response)
			if err != nil {
				status, body = InternalServerErrorResponse()
				write(status, body)
				s.Logger.Error(err)
				return
			}
		}
		// Get Response from model
		status = response.Status
		headers := response.Headers
		// Encode the response body to the appropriate format
		body, _ = model.Encode(response.Body)
		// Set response headers
		for key, value := range headers {
			w.Header().Set(key, value)
		}
		// Send response to client
		write(status, body)
		// if metrics client is defined, count this function call
		if s.Metrics != nil {
			s.Metrics.Incr(event, 1)
		}
	}
}

// Insert creates a http handler that will create a document in model's database.
func (s *Service) Insert(modelFactory ModelFactory) router.Handler {
	return s.persist(modelFactory, "insert")
}

// Update creates a http handler that will updates a document by the model's update selector in model's database
func (s *Service) Update(modelFactory ModelFactory) router.Handler {
	return s.persist(modelFactory, "update")
}

// Upsert creates a http handler that will upsert(create or update if it exists) a document selected by the model's upsert selector
func (s *Service) Upsert(modelFactory ModelFactory) router.Handler {
	return s.persist(modelFactory, "upsert")
}

// find creates a http handler that will list documents or return one document from a model's database
func (s *Service) find(modelFactory ModelFactory, mode string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// model is request scoped
		model := modelFactory.New(r)
		// event is the name used to track metrics
		event := model.Name() + "_find_" + mode
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// Send response back to client
		write := func(status int, body []byte) {
			w.WriteHeader(status)
			w.Write(body)
		}
		// Validate user input
		verr := model.Validate()
		if verr != nil {
			status, body = verr.Code, []byte(verr.Message)
			write(status, body)
			s.Logger.Error(verr.Message)
			return
		}
		var err error
		switch {
		case mode == "one":
			err = model.FindOne()
		case mode == "many":
			err = model.FindMany()
		}
		if err != nil {
			// Something wicked happened while fetching document/s
			status, body = InternalServerErrorResponse()
			write(status, body)
			s.Logger.Error(err)
			return
		}
		response := model.Response()
		// Notify other services, if an event broker exists
		if s.Broker != nil {
			err = s.Broker.Publish(event, response)
			if err != nil {
				status, body = InternalServerErrorResponse()
				write(status, body)
				s.Logger.Error(err)
				return
			}
		}
		body, _ = model.Encode(response.Body)
		status = response.Status
		// Set response headers
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		write(status, body)
		// If a metrics client is defined count this successful request
		if s.Metrics != nil {
			s.Metrics.Incr(event, 1)
		}
	}
}

// FindOne - creates a http handler that will return one document from a model's database if the id exists
func (s *Service) FindOne(modelFactory ModelFactory) router.Handler {
	return s.find(modelFactory, "one")
}

// FindMany - creates a http handler that will list documents from a model's database
func (s *Service) FindMany(modelFactory ModelFactory) router.Handler {
	return s.find(modelFactory, "many")
}

// Remove creates a http handler that will delete a document by remove selector specified in the model model
func (s *Service) Remove(modelFactory ModelFactory) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		model := modelFactory.New(r)
		// event is the name used to track metrics
		event := model.Name() + "_delete"
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Send response back to client
		write := func(status int, body []byte) {
			w.WriteHeader(status)
			w.Write(body)
		}
		// Validate the user input
		verr := model.Validate()
		if verr != nil {
			status, body = verr.Code, []byte(verr.Message)
			s.Logger.Error(verr.Message)
			write(status, body)
			return
		}
		// Remove the item if it exists
		err := model.Remove()
		if err != nil {
			status, body = InternalServerErrorResponse()
			write(status, body)
			s.Logger.Error(err)
			return
		}
		response := model.Response()
		if s.Broker != nil {
			err = s.Broker.Publish(event, response)
			if err != nil {
				status, body = InternalServerErrorResponse()
				write(status, body)
				s.Logger.Error(err)
				return
			}
		}
		// Set response headers
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		status = response.Status
		body, _ = model.Encode(response.Body)
		// Notify other services, if an event broker exists
		write(status, body)
		// If a metrics client is defined count this successful request
		if s.Metrics != nil {
			s.Metrics.Incr(event, 1)
		}
	}
}
