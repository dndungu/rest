package rest

import (
	"gopkg.in/zatiti/router.v1"
	"net/http"
)

// write creates a http handler for creating or updating a document depending on the mode provided
func (s *Service) write(modelFactory ModelFactory, mode string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// model is request scoped
		model := modelFactory.New(r)
		// event is the name used to track metrics
		event := model.Name() + "_" + mode + "_one"
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Instanciate a value of the model being created from the request body
		err := model.Decode()
		if err == nil {
			// Validate what came through the wire
			verr := model.Validate()
			if verr == nil {
				var v interface{}
				if mode == "insert" {
					v, err = model.Create()
					status = http.StatusCreated
				}
				if mode == "update" {
					v, err = model.Update()
					status = http.StatusNoContent
				}
				if mode == "upsert" {
					v, err = model.Upsert()
					status = http.StatusAccepted
				}
				if err == nil {
					// The output from model.Create could be invalid
					body, _ = model.Encode(v)
					//if err == nil {
					// Allow event broker to be optional
					if s.Broker != nil {
						err = s.Broker.Publish(event, body)
					}
					if err == nil {
						// If a metrics client is defined use it
						if s.Metrics != nil {
							s.Metrics.Incr(event, 1)
						}
					} else {
						// Something wicked happened while publishing to the event stream
						s.Logger.Error(err)
						status, body = InternalServerErrorResponse()
					}
					//} else {
					//	s.Logger.Error(err)
					//	status, body = InternalServerErrorResponse()
					//}
				} else {
					s.Logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
			} else {
				s.Logger.Error(verr.Message)
				status, body = verr.Code, []byte(verr.Message)
			}
		} else {
			s.Logger.Error(err)
			status, body = BadRequestResponse()
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Insert creates a http handler that will create a document in model's database.
func (s *Service) Insert(modelFactory ModelFactory) router.Handler {
	return s.write(modelFactory, "insert")
}

// Update creates a http handler that will updates a document by the model's update selector in model's database
func (s *Service) Update(modelFactory ModelFactory) router.Handler {
	return s.write(modelFactory, "update")
}

// Upsert creates a http handler that will upsert(create or update if it exists) a document selected by the model's upsert selector
func (s *Service) Upsert(modelFactory ModelFactory) router.Handler {
	return s.write(modelFactory, "upsert")
}

// find creates a http handler that will list documents or return one document from a model's database
func (s *Service) find(modelFactory ModelFactory, mode string) router.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		// model is request scoped
		model := modelFactory.New(r)
		// event is the name used to track metrics
		event := model.Name() + "_find_many"
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// Validate
		verr := model.Validate()
		if verr == nil {
			var v interface{}
			var err error
			if mode == "one" {
				v, err = model.FindOne()
			}
			if mode == "many" {
				v, err = model.FindMany()
			}
			if err == nil {
				body, _ = model.Encode(v)
				// Notify other services, if an event broker exists
				if s.Broker != nil {
					err = s.Broker.Publish(event, body)
				}
				if err == nil {
					status = http.StatusOK
					if s.Metrics != nil {
						s.Metrics.Incr(event, 1)
					}
				} else {
					s.Logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
			} else {
				// Something wicked happened while fetching document/s
				s.Logger.Error(err)
				status, body = InternalServerErrorResponse()
			}
		} else {
			s.Logger.Error(verr.Message)
			status, body = verr.Code, []byte(verr.Message)
		}
		w.WriteHeader(status)
		w.Write(body)
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
		event := model.Name() + "_delete_one"
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		verr := model.Validate()
		if verr == nil {
			// Remove the item if it exists
			v, err := model.Remove()
			if err == nil {
				// Encode the output into []byte
				body, _ = model.Encode(v)
				//if err == nil {
				// Notify other services, if an event broker exists
				if s.Broker != nil {
					err = s.Broker.Publish(event, body)
				}
				if err == nil {
					status, body = NoContentResponse()
					// If a metrics client is defined use it
					if s.Metrics != nil {
						s.Metrics.Incr(event, 1)
					}
				} else {
					status, body = InternalServerErrorResponse()
					s.Logger.Error(err)
				}
				//} else {
				//	status, body = InternalServerErrorResponse()
				//	s.Logger.Error(err)
				//}
			} else {
				status, body = InternalServerErrorResponse()
				s.Logger.Error(err)
			}
		} else {
			status, body = verr.Code, []byte(verr.Message)
			s.Logger.Error(verr.Message)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}
