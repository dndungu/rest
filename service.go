package rest

// Service holds application scope broker, logger and metrics adapters
type Service struct {
	broker  Broker
	logger  Logger
	metrics Metrics
}

// Broker is an event stream adapter to notify other microservices of state changes
type Broker interface {
	Publish(event string, v interface{}) error
}

// Logger is an leveled logging interface
type Logger interface {
	Info(v interface{})
	Warning(v interface{})
	Error(v interface{})
	Fatal(v interface{})
}

// Metrics is an adapter to track application performance metrics
type Metrics interface {
	Incr(stat string, count int64)
	Timing(stat string, delta int64)
	NewTimer(stat string) func()
}
