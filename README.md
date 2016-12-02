[![GoDoc](https://godoc.org/github.com/dndungu/rest?status.svg)](https://godoc.org/github.com/dndungu/rest)
[![Build Status](https://travis-ci.org/dndungu/rest.svg?branch=master)](https://travis-ci.org/dndungu/rest)
[![codecov](https://codecov.io/gh/dndungu/rest/branch/master/graph/badge.svg)](https://codecov.io/gh/dndungu/rest)
[![Go Report Card](https://goreportcard.com/badge/github.com/dndungu/rest)](https://goreportcard.com/report/github.com/dndungu/rest)
[![Issue Count](https://codeclimate.com/github/dndungu/rest/badges/issue_count.svg)](https://codeclimate.com/github/dndungu/rest)

# REST
REST is an opinionated library for quickly creating RESTful micro services. It is built using lessons learned while architecting micro services.

## Event driven architecture
REST makes it easy to use an event broker to send state changes between services.

## Metrics
REST makes it easy to track function performance metrics.

## Unit Testing
REST makes it easy to mock database, metrics client, and event broker to allow for 100% test code coverage in a RESTful.

## Example

```go
    f := NewFactory("todo").
        SetDefaultHeaders(headers).
        UseType(reflect.TypeOf(FakeFields{})).
        UseStorage(&FakeStorage{fail: s.failDB}).
        UseValidator(&FakeValidator{}).
        UseSerializer(&JSON{})

```
