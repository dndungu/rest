package rest

import (
	"os"
	"runtime"
	"time"
)

const (
	INFO    = "info"
	WARNING = "warning"
	ERROR   = "error"
	FATAL   = "fatal"
)

type Log struct {
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
	Hostname  string    `json:"hostname"`
	File      string    `json:"file"`
	Severity  string    `json:"severity"`
}

type LoggingSink interface {
	Write(l *Log)
}

type LoggingClient struct {
	Sink *[]LoggingSink
}

func (lc *LoggingClient) UseSinks(ls *[]LoggingSink) {
	lc.Sink = ls
}

func (lc *LoggingClient) CreateLog(s string, e error) (l *Log) {
	l = &Log{}
	l.CreatedAt = time.Now().UTC()
	l.Message = e.Error()
	l.Hostname, _ = os.Hostname()
	_, l.File, _, _ = runtime.Caller(4)
	l.Severity = s
	return l
}

func (lc *LoggingClient) Log(s string, e error) {
	l := lc.CreateLog(s, e)
	for _, sink := range *lc.Sink {
		sink.Write(l)
	}
}

func (lc *LoggingClient) Info(e error) {
	lc.Log(INFO, e)
}
func (lc *LoggingClient) Warning(e error) {
	lc.Log(WARNING, e)
}
func (lc *LoggingClient) Error(e error) {
	lc.Log(ERROR, e)
}
func (lc *LoggingClient) Fatal(e error) {
	lc.Log(FATAL, e)
	os.Exit(1)
}
