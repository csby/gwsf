package gdoc

import (
	"github.com/csby/gwsf/gtype"
	"sort"
	"strings"
)

type Output struct {
	Headers []*Header       `json:"headers"` // 头部
	Model   []*Type         `json:"model"`   // 数据模型
	Example interface{}     `json:"example"` // 数据示例
	Format  int             `json:"format"`  // 数据格式: 0-text; 1-json; 2-xml
	Errors  ErrorCollection `json:"errors"`  // 输出错误代码
}

func (s *Output) GetHeader(name string) *Header {
	c := len(s.Headers)
	for i := 0; i < c; i++ {
		item := s.Headers[i]
		if item == nil {
			continue
		}
		if strings.ToLower(item.Name) == strings.ToLower(name) {
			return item
		}
	}

	return nil
}

func (s *Output) RemoveHeader(name string) {
	items := make([]*Header, 0)
	c := len(s.Headers)
	for i := 0; i < c; i++ {
		item := s.Headers[i]
		if item == nil {
			continue
		}
		if strings.ToLower(item.Name) == strings.ToLower(name) {
			continue
		}
		items = append(items, item)
	}

	s.Headers = items
}

func (s *Output) ClearHeaders() {
	s.Headers = make([]*Header, 0)
}

func (s *Output) AddError(err gtype.Error) {
	if err == nil {
		return
	}

	s.AddErrorCustom(err.Code(), err.Summary())
}

func (s *Output) AddErrorCustom(code int, summary string) {
	count := len(s.Errors)
	for index := 0; index < count; index++ {
		item := s.Errors[index]
		if item.Code == code {
			item.Summary = summary
			return
		}
	}

	s.Errors = append(s.Errors, &Error{Code: code, Summary: summary})
	sort.Sort(s.Errors)
}

func (s *Output) ClearErrors() {
	s.Errors = make(ErrorCollection, 0)
}
