package gdoc

import "strings"

type Input struct {
	Headers []*Header   `json:"headers"` // 头部
	Queries []*Query    `json:"queries"` // 参数
	Forms   []*Form     `json:"forms"`   // 表单
	Model   []*Type     `json:"model"`   // 数据模型
	Example interface{} `json:"example"` // 数据示例
}

func (s *Input) GetHeader(name string) *Header {
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

func (s *Input) RemoveHeader(name string) {
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

func (s *Input) ClearHeaders() {
	s.Headers = make([]*Header, 0)
}

func (s *Input) GetQuery(name string) *Query {
	c := len(s.Queries)
	for i := 0; i < c; i++ {
		item := s.Queries[i]
		if item == nil {
			continue
		}
		if strings.ToLower(item.Name) == strings.ToLower(name) {
			return item
		}
	}

	return nil
}

func (s *Input) RemoveQueries(name string) {
	items := make([]*Query, 0)
	c := len(s.Queries)
	for i := 0; i < c; i++ {
		item := s.Queries[i]
		if item == nil {
			continue
		}
		if strings.ToLower(item.Name) == strings.ToLower(name) {
			continue
		}
		items = append(items, item)
	}

	s.Queries = items
}

func (s *Input) ClearQuery() {
	s.Queries = make([]*Query, 0)
}

func (s *Input) GetForm(key string) *Form {
	c := len(s.Forms)
	for i := 0; i < c; i++ {
		item := s.Forms[i]
		if item == nil {
			continue
		}
		if strings.ToLower(item.Key) == strings.ToLower(key) {
			return item
		}
	}

	return nil
}

func (s *Input) RemoveForm(key string) {
	items := make([]*Form, 0)
	c := len(s.Forms)
	for i := 0; i < c; i++ {
		item := s.Forms[i]
		if item == nil {
			continue
		}
		if strings.ToLower(item.Key) == strings.ToLower(key) {
			continue
		}
		items = append(items, item)
	}

	s.Forms = items
}

func (s *Input) ClearForms() {
	s.Forms = make([]*Form, 0)
}
