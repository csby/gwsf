package gtype

import (
	"fmt"
	"strings"
)

var (
	ErrSuccess    = newError(0, "")
	ErrUnknown    = newError(1, "未知错误")
	ErrInternal   = newError(2, "内部错误")
	ErrException  = newError(3, "内部异常")
	ErrNotSupport = newError(4, "不支持的操作")
	ErrExist      = newError(5, "已存在")
	ErrNotExist   = newError(6, "不存在")
	ErrInput      = newError(7, "输入错误")

	ErrTokenEmpty   = newError(101, "缺少凭证")
	ErrTokenInvalid = newError(101, "凭证无效")
	ErrTokenIllegal = newError(101, "凭证非法")

	ErrLoginCaptchaInvalid           = newError(201, "验证码无效")
	ErrLoginAccountNotExit           = newError(202, "账号不存在")
	ErrLoginPasswordInvalid          = newError(203, "密码不正确")
	ErrLoginAccountOrPasswordInvalid = newError(204, "账号或密码不正确")

	ErrNoPermission = newError(301, "没有权限")
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
		detail:  s.toString(v),
	}

	return err
}

func (s *innerError) toString(a []interface{}) string {
	sb := &strings.Builder{}
	c := len(a)
	for i := 0; i < c; i++ {
		item := a[i]
		if item == nil {
			continue
		}
		sb.WriteString(fmt.Sprint(item))
	}

	return sb.String()
}
