package rest

import (
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

const (
	INFO    = "info"
	WARNING = "warning"
	ERROR   = "error"
	FATAL   = "fatal"
)

type Log struct {
	CreatedAt  time.Time `json:"created_at"`
	Details    string    `json:"details"`
	StackTrace string    `json:"stack_trace"`
	Hostname   string    `json:"hostname"`
	File       string    `json:"file"`
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

func (lc *LoggingClient) CreateLog(e error) (l *Log) {
	l = &Log{}
	l.CreatedAt = time.Now().UTC()
	l.Details = e.Error()
	l.StackTrace = string(debug.Stack())
	l.Hostname, _ = os.Hostname()
	_, file, line, _ := runtime.Caller(2)
	l.File = file + ":" + strconv.Itoa(line)
	return l
}

func (lc *LoggingClient) Error(e error) {
	l := lc.CreateLog(e)
	for _, sink := range *lc.Sink {
		sink.Write(l)
	}
}
