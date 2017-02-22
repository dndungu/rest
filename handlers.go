package rest

import (
	"gopkg.in/zatiti/router.v1"
	"net/http"
)

// rovided
func (s *Service) process(resource *Resource, action string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		var response Response
		// When a new request comes in we want a new model instance created to handle that request.
		model := resource.NewModel(r, action)
		// Event is the name used to track the transaction,
		event := model.Name + "_" + action
		// Track how long this function take to return.
		stop := s.Metrics.NewTimer(event)
		// Send response back to client when this function returns
		defer func() {
			// Get a pointer to the response struct
			response = model.GetResponse()
			status := response.Status
			body, err := model.Encode(response.Body)
			if err != nil {
				status = http.StatusInternalServerError
				body = []byte(http.StatusText(http.StatusInternalServerError))
				s.Logger.Error(err)
			}
			// Set response headers
			for key, value := range response.Headers {
				w.Header().Set(key, value[0])
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
			s.Logger.Error(err)
			return
		}
		// Execute database operation
		err = model.Execute(action)
		// Handle failed database operation
		if err != nil {
			s.Logger.Error(err)
		}
		// If event broker is defined send the event to through the stream
		err = s.Broker.Publish(event, &Event{Request: r, Response: &response})
		if err != nil {
			model.SetResponseStatus(http.StatusInternalServerError)
			s.Logger.Error(err)
		}
		err = s.Metrics.Incr(event, 1)
		if err != nil {
			model.SetResponseStatus(http.StatusInternalServerError)
			s.Logger.Error(err)
		}
		stop()
	}
}

// InsertOne creates a http handler that will create a document in model's database.
func (s *Service) InsertOne(resource *Resource) router.Handler {
	return s.process(resource, "insertOne")
}

// InsertMany creates a http handler that will create a document in model's database.
func (s *Service) InsertMany(resource *Resource) router.Handler {
	return s.process(resource, "insertMany")
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
	return s.process(resource, "findOne")
}

// FindMany - creates a http handler that will list documents from a model's database
func (s *Service) FindMany(resource *Resource) router.Handler {
	return s.process(resource, "findMany")
}

// Remove creates a http handler that will delete a document by remove selector specified in the model model
func (s *Service) Remove(resource *Resource) router.Handler {
	return s.process(resource, "remove")
}
