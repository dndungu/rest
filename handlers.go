package rest

import (
	"gopkg.in/zatiti/router.v1"
	"net/http"
)

// rovided
func (s *Service) process(resource *Resource, action string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// When a new request comes in we want a new model instance created to handle that request.
		model := resource.New(r, action)
		// Event is the name used to track the transaction,
		event := model.Name + "_" + action
		// Track how long this function take to return.
		stop := s.Metrics.NewTimer(event)
		// Send response back to client when this function returns
		defer func() {
			// Get a pointer to the response struct
			response := model.Context.Response
			status := response.Status
			body, err := model.Encode(response.Body)
			if err != nil {
				status = http.StatusInternalServerError
				body = []byte(http.StatusText(http.StatusInternalServerError))
				s.Logger.Error(err)
			}
			// Set response headers
			for key, value := range response.Headers {
				w.Header().Set(key, value)
			}
			// Write the response status code
			w.WriteHeader(status)
			// Write the response body
			w.Write(body)
		}()
		var err error
		err = model.Decode()
		if err != nil {
			s.Logger.Error(err)
			return
		}
		// Validate user input
		err = model.Validate()
		if err != nil {
			s.Logger.Error(err.Error())
			return
		}
		switch {
		case action == "insert_one":
			err = model.InsertOne()
		case action == "insert_many":
			err = model.InsertMany()
		case action == "update":
			err = model.Update()
		case action == "upsert":
			err = model.Upsert()
		case action == "find_one":
			err = model.FindOne()
		case action == "find_many":
			err = model.FindMany()
		case action == "remove":
			err = model.Remove()
		}
		// Handle failed database operation
		if err != nil {
			s.Logger.Error(err)
		}
		// If event broker is defined send the event to through the stream
		err = s.Broker.Publish(event, &Event{Request: r, Response: &model.Context.Response})
		if err != nil {
			model.Context.Response.Status = http.StatusInternalServerError
			s.Logger.Error(err)
		}
		err = s.Metrics.Incr(event, 1)
		if err != nil {
			model.Context.Response.Status = http.StatusInternalServerError
			s.Logger.Error(err)
		}
		stop()
	}
}

// InsertOne creates a http handler that will create a document in model's database.
func (s *Service) InsertOne(resource *Resource) router.Handler {
	return s.process(resource, "insert_one")
}

// InsertMany creates a http handler that will create a document in model's database.
func (s *Service) InsertMany(resource *Resource) router.Handler {
	return s.process(resource, "insert_many")
}

// Update creates a http handler that will updates a document by the model's update selector in model's database
func (s *Service) Update(resource *Resource) router.Handler {
	return s.process(resource, "update")
}

// Upsert creates a http handler that will upsert(create or update if it exists) a document selected by the model's upsert selector
func (s *Service) Upsert(resource *Resource) router.Handler {
	return s.process(resource, "upsert")
}

// FindOne - creates a http handler that will return one document from a model's database if the id exists
func (s *Service) FindOne(resource *Resource) router.Handler {
	return s.process(resource, "find_one")
}

// FindMany - creates a http handler that will list documents from a model's database
func (s *Service) FindMany(resource *Resource) router.Handler {
	return s.process(resource, "find_many")
}

// Remove creates a http handler that will delete a document by remove selector specified in the model model
func (s *Service) Remove(resource *Resource) router.Handler {
	return s.process(resource, "remove")
}
