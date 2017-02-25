package rest

// Service holds application scope broker, logger and metrics adapters
type Service struct {
	Broker  Broker
	Logger  Logger
	Metrics Metrics
}

// Broker is an event stream adapter to notify other microservices of state changes
type Broker interface {
	Publish(event string, v *Event) error
}

// Logger is an leveled logging interface
type Logger interface {
	Info(v error)
	Warning(v error)
	Error(v error)
	Fatal(v error)
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
	if s.Metrics == nil {
		return func() {}
	}
	return s.Metrics.NewTimer(stat)
}
