package gtype

import "net/http"

type Context interface {
	Request() *http.Request
	Response() http.ResponseWriter
	Query(name string) string
	GetBody() ([]byte, error)
	GetJson(v interface{}) error
	GetXml(v interface{}) error
	OutputJson(v interface{})
	OutputXml(v interface{})
	Success(data interface{})
	Error(err Error, detail ...interface{})

	Method() string
	Schema() string
	Host() string
	Path() string
	Certificate() *Certificate
	SetHandled(v bool)
	IsHandled() bool
	Token() string
	RID() uint64
	RIP() string
	NewGuid() string
}
