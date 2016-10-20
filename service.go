package rest

// Service holds application scope broker, logger and metrics adapters
type Service struct {
	broker     Broker
	logger     Logger
	metrics    Metrics
	serializer Serializer
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

// NewTimer creates a stop timer to track the performance of a function
func (s *Service) NewTimer(stat string) func() {
	// Allow metrics to be optional
	if s.metrics == nil {
		return func() {}
	}
	return s.metrics.NewTimer(stat)
}
