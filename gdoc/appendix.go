package gdoc

import "github.com/csby/gwsf/gtype"

type Appendix struct {
	Label string          `json:"label"` // 标签
	Items []*AppendixItem `json:"items"` // 项目
}

func (s *Appendix) Add(value interface{}, name, note string, example interface{}) {
	if s.Items == nil {
		s.Items = make([]*AppendixItem, 0)
	}

	item := &AppendixItem{
		Value:   value,
		Name:    name,
		Note:    note,
		Example: example,
	}
	arg := modelArgument.FromExample(example)
	if arg != nil {
		item.Model = arg.ToModel()
	} else {
		item.Model = make([]*Type, 0)
	}

	s.Items = append(s.Items, item)
}

func (s *Appendix) AddItem(item gtype.AppendixItem) {
	if item != nil {
		s.Add(item.Value(), item.Name(), item.Note(), item.Example())
	}
}

type AppendixItem struct {
	Value   interface{} `json:"value"`   // 值
	Name    string      `json:"name"`    // 名称
	Note    string      `json:"note"`    // 备注
	Model   []*Type     `json:"model"`   // 数据模型
	Example interface{} `json:"example"` // 数据示例
}
