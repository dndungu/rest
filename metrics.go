package rest

import (
	"time"
)

// MetricsClient - satisfies a statsd interface
type MetricsClient interface {
	Incr(string, []string, float64) error
	Timing(string, time.Duration, []string, float64) error
}

// ServiceMetrics - a metrics client
type ServiceMetrics struct {
	Client MetricsClient
	Logger Logger
	Tags   []string
}

// NewServiceMetrics - creates and returns a new Metrics instance
func NewServiceMetrics() *ServiceMetrics {
	return &ServiceMetrics{}
}

// UseClient -
func (sm *ServiceMetrics) UseClient(client MetricsClient) *ServiceMetrics {
	sm.Client = client
	return sm
}

// UseLogger -
func (sm *ServiceMetrics) UseLogger(logger Logger) *ServiceMetrics {
	sm.Logger = logger
	return sm
}

// UseTags -
func (sm *ServiceMetrics) UseTags(tags []string) *ServiceMetrics {
	sm.Tags = tags
	return sm
}

// Incr - record an increment by count
func (sm *ServiceMetrics) Incr(stat string, count int64) error {
	err := sm.Client.Incr(stat, sm.Tags, float64(count))
	if err != nil {
		sm.Logger.Error(err)
	}
	return err
}

// Timing - record the time taken to complete an operation
func (sm *ServiceMetrics) Timing(stat string, delta int64) error {
	err := sm.Client.Timing(stat, time.Duration(delta), sm.Tags, 1)
	if err != nil {
		sm.Logger.Error(err)
	}
	return err
}

// NewTimer - create a function that will calculate and record the timing when called
func (sm *ServiceMetrics) NewTimer(stat string) func() {
	start := time.Now()
	return func() {
		delta := time.Since(start)
		sm.Timing(stat, int64(delta))
	}
}
