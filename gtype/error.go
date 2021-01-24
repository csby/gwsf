package gtype

import "fmt"

var (
	ErrSuccess    = newError(0, "")
	ErrUnknown    = newError(1, "未知错误")
	ErrInternal   = newError(2, "内部错误")
	ErrException  = newError(3, "内部异常")
	ErrNotSupport = newError(4, "不支持的操作")
	ErrExist      = newError(5, "已存在")
	ErrNotExist   = newError(6, "不存在")
	ErrInput      = newError(7, "输入错误")
)

type Error interface {
	Code() int
	Summary() string
	Detail() string
	Inner() error
	New(inner error) Error
	SetDetail(v ...interface{}) Error
}

func NewError(code int, summary string, inner error, details ...interface{}) Error {
	err := &innerError{
		code:    code,
		summary: summary,
		inner:   inner,
		detail:  fmt.Sprint(details...),
	}

	if inner != nil && len(err.detail) < 1 {
		err.detail = inner.Error()
	}

	return err
}

func newError(code int, summary string) Error {
	return &innerError{
		code:    code,
		summary: summary,
		detail:  "",
		inner:   nil,
	}
}

type innerError struct {
	code    int
	summary string
	detail  string
	inner   error
}

func (s *innerError) Code() int {
	return s.code
}
func (s *innerError) Summary() string {
	return s.summary
}
func (s *innerError) Detail() string {
	return s.detail
}
func (s *innerError) Inner() error {
	return s.inner
}
func (s *innerError) New(inner error) Error {
	err := &innerError{
		code:    s.code,
		summary: s.summary,
		inner:   inner,
	}

	if inner != nil {
		err.detail = inner.Error()
	}

	return err
}
func (s *innerError) SetDetail(v ...interface{}) Error {
	err := &innerError{
		code:    s.code,
		summary: s.summary,
		detail:  fmt.Sprint(v...),
	}

	return err
}
