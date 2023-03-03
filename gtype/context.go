package gtype

import (
	"net/http"
	"time"
)

const (
	CtxUserAccount = "ctx_user_account"
)

type Context interface {
	Request() *http.Request
	Response() http.ResponseWriter
	Query(name string) string
	GetBody() ([]byte, error)
	GetJson(v interface{}) error
	GetXml(v interface{}) error
	GetSoapAction() string
	OutputJson(v interface{})
	OutputXml(v interface{})
	OutputSoap(v interface{})
	Success(data interface{})
	Error(err Error, detail ...interface{})
	ErrorWithData(data interface{}, err Error, detail ...interface{})

	EnterTime() time.Time
	LeaveTime() *time.Time
	Method() string
	Schema() string
	Host() string
	Path() string
	Queries() QueryCollection
	Certificate() *Certificate
	SetHandled(v bool)
	IsHandled() bool
	Token() string
	Node() string
	Instance() string
	ForwardFrom() string
	RID() uint64
	RIP() string
	NewGuid() string
	GetInput() []byte
	GetInputFormat() int
	GetOutput() []byte
	GetOutputFormat() int
	GetOutputCode() *int
	IsError() bool
	SetLog(v bool)
	GetLog() bool

	Set(key string, val interface{})
	Get(key string) (interface{}, bool)
	Del(key string) bool

	ClientOrganization() string
	SetClientOrganization(ou string)
	FireAfterInput()
}
