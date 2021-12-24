package gdoc

import (
	"encoding/xml"
	"github.com/csby/gwsf/gtype"
	"strings"
)

var (
	modelArgument = &argument{}
)

type Function struct {
	ID          string  `json:"id"`          // 接口标识
	Name        string  `json:"name"`        // 接口名称
	Note        string  `json:"note"`        // 接口说明
	Remark      string  `json:"remark"`      // 接口备注
	Method      string  `json:"method"`      // 接口方法
	Path        string  `json:"path"`        // 接口地址
	FullPath    string  `json:"fullPath"`    // 接口地址
	IsWebsocket bool    `json:"isWebsocket"` // 是否为websocket接口
	Input       *Input  `json:"input"`       // 输入
	Output      *Output `json:"output"`      // 输出

	TokenUI     func() []gtype.TokenUI                                                 `json:"-"`
	TokenCreate func(items []gtype.TokenAuth, ctx gtype.Context) (string, gtype.Error) `json:"-"`
}

func (s *Function) SetNote(v string) {
	s.Note = v
}

func (s *Function) SetRemark(v string) {
	s.Remark = v
}

func (s *Function) AddInputHeader(required bool, name, note, defaultValue string, optionValues ...string) {
	header := s.Input.GetHeader(name)
	if header != nil {
		header.Required = required
		header.Note = note
		header.DefaultValue = defaultValue
		header.Values = optionValues
		header.Token = strings.EqualFold(name, gtype.TokenName)
	} else {
		s.Input.Headers = append(s.Input.Headers, &Header{
			Name:         name,
			Note:         note,
			Required:     required,
			Values:       optionValues,
			DefaultValue: defaultValue,
			Token:        strings.EqualFold(name, gtype.TokenName),
		})
	}
}

func (s *Function) ClearInputHeader() {
	s.Input.ClearHeaders()
}

func (s *Function) RemoveInputHeader(name string) {
	s.Input.RemoveHeader(name)
}

func (s *Function) AddInputQuery(required bool, name, note, defaultValue string, optionValues ...string) {
	query := s.GetInputQuery(name)
	if query != nil {
		query.Required = required
		query.Note = note
		query.DefaultValue = defaultValue
		query.Values = optionValues
		query.Token = strings.EqualFold(name, gtype.TokenName)
	} else {
		s.Input.Queries = append(s.Input.Queries, &Query{
			Name:         name,
			Note:         note,
			Required:     required,
			Values:       optionValues,
			DefaultValue: defaultValue,
			Token:        strings.EqualFold(name, gtype.TokenName),
		})
	}
}

func (s *Function) RemoveInputQuery(name string) {
	s.Input.RemoveQueries(name)
}

func (s *Function) AddInputForm(required bool, key, note string, valueKind int, defaultValue interface{}) {
	form := s.GetInputForm(key)
	if form != nil {
		form.Required = required
		form.Note = note
		form.Value = defaultValue
		form.ValueKind = valueKind
	} else {
		s.Input.Forms = append(s.Input.Forms, &Form{
			Key:       key,
			Note:      note,
			Required:  required,
			Value:     defaultValue,
			ValueKind: valueKind,
		})
	}
}

func (s *Function) RemoveInputForm(key string) {
	s.Input.RemoveForm(key)
}

func (s *Function) SetInputFormat(v int) {
	s.Input.Format = v
}

func (s *Function) SetInputExample(v interface{}) {
	s.Input.Example = v
	arg := modelArgument.FromExample(v)
	if arg != nil {
		s.Input.Model = arg.ToModel()
	} else {
		s.Input.Model = make([]*Type, 0)
	}
}

func (s *Function) SetInputJsonExample(v interface{}) {
	s.SetInputFormat(gtype.ArgsFmtJson)
	s.SetInputExample(v)
	s.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeJson, gtype.ContentTypeJson)
}

func (s *Function) SetInputXmlExample(v interface{}) {
	s.SetInputFormat(gtype.ArgsFmtXml)
	example, err := xml.MarshalIndent(v, "", "	")
	if err == nil {
		s.Input.Example = string(example)
	}
	s.AddInputHeader(true, "content-type", "内容类型", gtype.ContentTypeXml, gtype.ContentTypeXml)
}

func (s *Function) AddOutputHeader(name, value string) {
	header := s.GetOutputHeader(name)
	if header != nil {
		header.DefaultValue = value
	} else {
		s.Output.Headers = append(s.Output.Headers, &Header{
			Name:         name,
			DefaultValue: value,
		})
	}
}

func (s *Function) ClearOutputHeader() {
	s.Output.ClearHeaders()
}

func (s *Function) AddOutputError(err gtype.Error) {
	if err == nil {
		return
	}
	s.AddOutputErrorCustom(err.Code(), err.Summary())
}

func (s *Function) AddOutputErrorCustom(code int, summary string) {
	s.Output.AddErrorCustom(code, summary)
}

func (s *Function) SetOutputFormat(v int) {
	s.Output.Format = v
}

func (s *Function) SetOutputExample(v interface{}) {
	s.Output.Example = v
	argument := modelArgument.FromExample(v)
	if argument != nil {
		s.Output.Model = argument.ToModel()
	} else {
		s.Output.Model = make([]*Type, 0)
	}
}

func (s *Function) SetOutputDataExample(v interface{}) {
	s.Output.Example = &gtype.Result{
		Code:   0,
		Serial: 201805161315480008,
		Data:   v,
	}
	argument := modelArgument.FromExample(s.Output.Example)
	if argument != nil {
		s.Output.Model = argument.ToModel()
	} else {
		s.Output.Model = make([]*Type, 0)
	}
	s.SetOutputFormat(gtype.ArgsFmtJson)
	s.AddOutputError(gtype.ErrException)
	s.AddOutputHeader("content-type", "application/json;charset=utf-8")
}

func (s *Function) SetOutputXmlExample(v interface{}) {
	s.SetOutputFormat(gtype.ArgsFmtXml)
	example, err := xml.MarshalIndent(v, "", "	")
	if err == nil {
		s.Output.Example = string(example)
	}
	s.AddOutputHeader("content-type", "application/xml;charset=utf-8")
}

func (s *Function) GetInputHeader(name string) *Header {
	return s.Input.GetHeader(name)
}

func (s *Function) GetInputQuery(name string) *Query {
	return s.Input.GetQuery(name)
}

func (s *Function) GetInputForm(key string) *Form {
	return s.Input.GetForm(key)
}

func (s *Function) GetOutputHeader(name string) *Header {
	return s.Output.GetHeader(name)
}

func (s *Function) tokenTypeChanged() {
}
