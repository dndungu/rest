package rest

import (
	"net/http"
)

// Insert creates a http handler that will create a document in mongodb.
// It takes a collection name and type struct of whitelisted fields
func (s *Service) Insert(modelFactory ModelFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// model is request scoped
		model := modelFactory.New(r)
		// event is the name used to track metrics
		event := model.Name() + "_insert_one"
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
			err = model.Validate()
			if err == nil {
				v, err := model.Create()
				if err == nil {
					body, _ = model.Encode(v)
					//if err == nil {
					// Allow event broker to be optional
					if s.broker != nil {
						err = s.broker.Publish(event, body)
					}
					if err == nil {
						status = http.StatusCreated
						// If a metrics client is defined use it
						if s.metrics != nil {
							s.metrics.Incr(event, 1)
						}
					} else {
						// Something wicked happened while publishing to the event stream
						s.logger.Error(err)
						status, body = InternalServerErrorResponse()
					}
				} else {
					s.logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
				//} else {
				// Something wicked happened while persisting document
				//	s.logger.Error(err)
				//	status, body = InternalServerErrorResponse()
				//}
			} else {
				s.logger.Error(err)
				status, body = BadRequestResponse()
			}
		} else {
			s.logger.Error(err)
			status, body = BadRequestResponse()
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Find creates a http handler that will list documents in mongodb
// It takes a collection name and type struct of fields to return
func (s *Service) Find(modelFactory ModelFactory) http.HandlerFunc {
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
		err := model.Validate()
		if err == nil {
			v, err := model.Find()
			if err == nil {
				body, _ = model.Encode(v)
				//if err == nil {
				// Notify other services, if an event broker exists
				if s.broker != nil {
					err = s.broker.Publish(event, body)
				}
				if err == nil {
					status = http.StatusOK
					if s.metrics != nil {
						s.metrics.Incr(event, 1)
					}
				} else {
					s.logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
				//	} else {
				//		s.logger.Error(err)
				//		status, body = InternalServerErrorResponse()
				//	}
			} else {
				// Something wicked happened while fetching document/s
				s.logger.Error(err)
				status, body = InternalServerErrorResponse()
			}
		} else {
			s.logger.Error(err)
			status, body = BadRequestResponse()
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Update creates a http handler that will updates a document by id in mongodb
// It takes a model creator argument
func (s *Service) Update(modelFactory ModelFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		model := modelFactory.New(r)
		// event is the name used to track metrics
		event := model.Name() + "_update_one"
		// Track how long this function take to return
		stop := s.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		err := model.Decode()
		if err == nil {
			err = model.Validate()
			if err == nil {
				v, err := model.Update()
				if err == nil {
					// Decode
					body, _ = model.Encode(v)
					//if err == nil {
					// Notify other services, if an event broker exists
					if s.broker != nil {
						err = s.broker.Publish(event, body)
					}
					if err == nil {
						status, body = NoContentResponse()
						// If a metrics client is defined use it
						if s.metrics != nil {
							s.metrics.Incr(event, 1)
						}
					} else {
						//Something wicked happened while trying to publish to the event strea,
						status, body = InternalServerErrorResponse()
						s.logger.Error(err)
					}
					//} else {
					//	status, body = InternalServerErrorResponse()
					//	s.logger.Error(err)
					//}
				} else {
					//Something wicked happened while persisting the update
					status, body = InternalServerErrorResponse()
					s.logger.Error(err)
				}

			} else {
				//The request failed validation rules in the model
				status, body = BadRequestResponse()
				s.logger.Error(err)
			}
		} else {
			status, body = BadRequestResponse()
			s.logger.Error(err)

		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Remove creates a http handler that will delete a document by id in mongodb
// It takes a model creator argument
func (s *Service) Remove(modelFactory ModelFactory) http.HandlerFunc {
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
		err := model.Validate()
		if err == nil {
			// Remove the item if it exists
			v, err := model.Remove()
			if err == nil {
				// Encode the output into []byte
				body, _ = model.Encode(v)
				//if err == nil {
				// Notify other services, if an event broker exists
				if s.broker != nil {
					err = s.broker.Publish(event, body)
				}
				if err == nil {
					status, body = NoContentResponse()
					// If a metrics client is defined use it
					if s.metrics != nil {
						s.metrics.Incr(event, 1)
					}
				} else {
					status, body = InternalServerErrorResponse()
					s.logger.Error(err)
				}
				//} else {
				//	status, body = InternalServerErrorResponse()
				//	s.logger.Error(err)
				//}
			} else {
				status, body = InternalServerErrorResponse()
				s.logger.Error(err)
			}
		} else {
			status, body = BadRequestResponse()
			s.logger.Error(err)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}
