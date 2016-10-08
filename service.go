package rest

type Service struct {
	broker  Broker
	logger  Logger
	metrics Metrics
}

type Broker interface {
	Publish(event string, v interface{}) error
}

type Logger interface {
	Info(v interface{})
	Warning(v interface{})
	Error(v interface{})
	Fatal(v interface{})
}

type Metrics interface {
	Incr(stat string, count int64)
	Timing(stat string, delta int64)
	NewTimer(stat string) func()
}
