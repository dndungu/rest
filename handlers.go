package rest

import (
	"gopkg.in/zatiti/router.v1"
	"net/http"
)

// write creates a http handler for creating or updating a document depending on the action provided
func (s *Service) process(modelFactory *ModelFactory, action string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// When a new request comes in we want a new model instance created to handle that request.
		model := modelFactory.New(r, action)
		// Event is the name used to track the transaction,
		event := model.Name + "_" + action
		// Track how long this function take to return.
		if s.Metrics != nil {
			stop := s.Metrics.NewTimer(event)
			defer func() {
				stop()
				err := s.Metrics.Incr(event, 1)
				if err != nil {
					s.Logger.Error(err)
				}
			}()
		}
		// Send response back to client when this function returns
		defer func() {
			// Get a pointer to the response struct
			response := model.Context.Response
			status := response.Status
			body, err := model.Encode(response.Body)
			if err != nil {
				body = []byte(http.StatusText(http.StatusInternalServerError))
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
		// Call the relevant model action
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
		if s.Broker != nil {
			// If event broker is defined send the event to through the stream
			err := s.Broker.Publish(event, &Event{Request: r, Response: &model.Context.Response})
			if err != nil {
				model.Context.Response.Status = http.StatusInternalServerError
				s.Logger.Error(err)
			}
		}
	}
}

// Insert creates a http handler that will create a document in model's database.
func (s *Service) InsertOne(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "insert_one")
}

// Insert creates a http handler that will create a document in model's database.
func (s *Service) InsertMany(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "insert_many")
}

// Update creates a http handler that will updates a document by the model's update selector in model's database
func (s *Service) Update(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "update")
}

// Upsert creates a http handler that will upsert(create or update if it exists) a document selected by the model's upsert selector
func (s *Service) Upsert(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "upsert")
}

// FindOne - creates a http handler that will return one document from a model's database if the id exists
func (s *Service) FindOne(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "find_one")
}

// FindMany - creates a http handler that will list documents from a model's database
func (s *Service) FindMany(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "find_many")
}

// Remove creates a http handler that will delete a document by remove selector specified in the model model
func (s *Service) Remove(modelFactory *ModelFactory) router.Handler {
	return s.process(modelFactory, "remove")
}
