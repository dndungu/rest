package rest

import (
	"encoding/json"
	"net/http"
)

type Model interface {
	Create() (interface{}, error)
	Find(r *http.Request) ([]interface{}, error)
	Validate() error
}

// Insert creates a http handler that will create a document in mongodb.
// It takes a collection name and type struct of whitelisted fields
func (service *Service) InsertOne(name string, model Model) http.HandlerFunc {
	event := name + ".insert.one"
	return func(w http.ResponseWriter, r *http.Request) {
		// Response StatusCode and Body
		var status int
		var body []byte
		// Allow metrics to be optional
		if service.metrics != nil {
			stop := service.metrics.NewTimer(event)
			defer stop()
		}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(model)
		// The body is not a valid JSON
		if err == nil {
			// Validate what came through the wire
			err = model.Validate()
			if err == nil {
				doc, err := model.Create()
				// Something wicked happened while persisting document
				if err == nil {
					body, _ = json.Marshal(doc)
					// Allow event broker to be optional
					if service.broker != nil {
						err = service.broker.Publish(event, doc)
					}
					// Something wicked happened while publishing to the event stream
					if err == nil {
						// Allow metrics to be optional
						if service.metrics != nil {
							service.metrics.Incr(event, 1)
						}
						status, body = CreatedResponse()
					} else {
						service.logger.Error(err)
						status, body = InternalServerErrorResponse()
					}
				} else {
					service.logger.Error(err)
					status, body = InternalServerErrorResponse()
				}
			} else {
				service.logger.Error(err)
				status, body = BadRequestResponse()
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
/*
func (service *Service) Find(name string, model Model) http.HandlerFunc {
	event := name + ".find.many"
	return func(w http.ResponseWriter, r *http.Request) {
		var status int
		var body []byte
		if service.metrics != nil {
			stop := service.metrics.NewTimer(event)
			defer stop()
		}
		models, err := model.Find(r)
		if err == nil {
			body, err = json.Marshal(models)
			if err == nil {
				service.metrics.Incr(event, 1)
				status = http.StatusOK
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

// FindOne creates a http handler that will find a document by id in mongodb
// It takes a collection name and type struct of fields to return
func FindOneById(name string, models []interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + ".findOne"
		stop := NewTimer(event)
		defer stop()
		c := db.C(name)
		id := router.Params(r).Get("id")
		err := c.FindId(id).All(&models)
		var status int
		var body []byte
		if fail(err) {
			status, body = InternalServerErrorResponse()
		} else {
			metrics.Incr(event, 1)
			body, err = json.Marshal(models)
			if fail(err) {
				status, body = InternalServerErrorResponse()
			} else {
				status = http.StatusOK
			}
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// DeleteOne creates a http handler that will delete a document by id in mongodb
// It takes a collection name
func DeleteOneById(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + ".deleteOne"
		stop := NewTimer(event)
		defer stop()
		c := db.C(name)
		params := router.Params(r)
		id := params.Get("id")
		err := c.Remove(bson.M{"id": id})
		var status int
		var body []byte
		if fail(err) {
			status, body = InternalServerErrorResponse()
		} else {
			status, body = NoContentResponse()
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// UpdateOne creates a http handler that will updates a document by id in mongodb
// It takes a collection name and type struct of whitelisted fields
func UpdateOne(name string, model interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := name + ".updateOne"
		stop := NewTimer(event)
		defer stop()
		c := db.C(name)
		params := router.Params(r)
		id := params.Get("id")
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&model)
		var status int
		var body []byte
		if fail(err) {
			status, body = BadRequestResponse()
		} else {
			err = c.Update(bson.M{"id": id}, model)
			if fail(err) {
				status, body = InternalServerErrorResponse()
			} else {
				status, body = NoContentResponse()
			}
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}
*/
