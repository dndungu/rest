package rest

import (
	"net/http"
)

// Insert creates a http handler that will create a document in mongodb.
// It takes a collection name and type struct of whitelisted fields
func (service *Service) Insert(name string, model Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + "_insert_one"
		// Track how long this function take to return
		stop := service.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Validate what came through the wire
		err := model.Validate(r)
		if err == nil {
			body, err = model.Create(r)
			if err == nil {
				// Allow event broker to be optional
				if service.broker != nil {
					err = service.broker.Publish(event, body)
				}
				if err == nil {
					status = http.StatusCreated
					// If a metrics client is defined use it
					if service.metrics != nil {
						service.metrics.Incr(event, 1)
					}
				} else {
					// Something wicked happened while publishing to the event stream
					service.logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
			} else {
				// Something wicked happened while persisting document
				service.logger.Error(err)
				status, body = InternalServerErrorResponse()
			}
		} else {
			service.logger.Error(err)
			status, body = BadRequestResponse()
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Find creates a http handler that will list documents in mongodb
// It takes a collection name and type struct of fields to return
func (service *Service) Find(name string, model Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + ".find.many"
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		// Track how long this function take to return
		stop := service.NewTimer(event)
		defer stop()
		err := model.Validate(r)
		if err == nil {
			body, err = model.Find(r)
			if err == nil {
				// Notify other services, if an event broker exists
				if service.broker != nil {
					err = service.broker.Publish(event, body)
				}
				if err == nil {
					status = http.StatusOK
					if service.metrics != nil {
						service.metrics.Incr(event, 1)
					}
				} else {
					service.logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
			} else {
				// Something wicked happened while fetching document/s
				service.logger.Error(err)
				status, body = InternalServerErrorResponse()
			}
		} else {
			service.logger.Error(err)
			status, body = BadRequestResponse()
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Delete creates a http handler that will delete a document by id in mongodb
// It takes a collection name
func (service *Service) Delete(name string, model Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + "_delete_one"
		// Track how long this function take to return
		stop := service.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		err := model.Validate(r)
		if err == nil {
			err := model.Delete(r)
			if err == nil {
				// Notify other services, if an event broker exists
				if service.broker != nil {
					err = service.broker.Publish(event, body)
				}
				if err == nil {
					status, body = NoContentResponse()
					// If a metrics client is defined use it
					if service.metrics != nil {
						service.metrics.Incr(event, 1)
					}
				} else {
					status, body = InternalServerErrorResponse()
					service.logger.Error(err)
				}
			} else {
				status, body = InternalServerErrorResponse()
				service.logger.Error(err)
			}
		} else {
			status, body = BadRequestResponse()
			service.logger.Error(err)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Update creates a http handler that will updates a document by id in mongodb
// It takes a collection name and type struct of whitelisted fields
func (service *Service) Update(name string, model Model) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + "_update_one"
		// Track how long this function take to return
		stop := service.NewTimer(event)
		defer stop()
		// HTTP response status code
		var status int
		// HTTP response body
		var body []byte
		err := model.Validate(r)
		if err == nil {
			body, err = model.Update(r)
			if err == nil {
				// Notify other services, if an event broker exists
				if service.broker != nil {
					err = service.broker.Publish(event, body)
				}
				if err == nil {
					status, body = NoContentResponse()
					// If a metrics client is defined use it
					if service.metrics != nil {
						service.metrics.Incr(event, 1)
					}
				} else {
					//Something wicked happened while trying to publish to the event strea,
					status, body = InternalServerErrorResponse()
					service.logger.Error(err)
				}
			} else {
				//Something wicked happened while persisting the update
				status, body = InternalServerErrorResponse()
				service.logger.Error(err)
			}

		} else {
			//The request failed validation rules in the model
			status, body = BadRequestResponse()
			service.logger.Error(err)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// NewTimer creates a stop timer to track the performance of a function
func (service *Service) NewTimer(stat string) func() {
	// Allow metrics to be optional
	if service.metrics == nil {
		return func() {}
	}
	return service.metrics.NewTimer(stat)
}
