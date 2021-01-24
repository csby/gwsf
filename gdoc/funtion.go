package gdoc

import (
	"github.com/csby/gwsf/gtype"
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
	TokenType   int     `json:"tokenType"`   // 凭证类型
	TokenPlace  int     `json:"tokenPlace"`  // 凭证位置: 0-header(头部); 1-query(URL参数)
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

func (s *Function) SetTokenType(v int) {
	if s.TokenType == v {
		return
	}
	s.TokenType = v
}

func (s *Function) SetInputContentType(v string) {

}

func (s *Function) AddInputHeader(required bool, name, note, defaultValue string, optionValues ...string) {
	header := s.Input.GetHeader(name)
	if header != nil {
		header.Required = required
		header.Note = note
		header.DefaultValue = defaultValue
		header.Values = optionValues
	} else {
		s.Input.Headers = append(s.Input.Headers, &Header{
			Name:         name,
			Note:         note,
			Required:     required,
			Values:       optionValues,
			DefaultValue: defaultValue,
		})
	}
}

func (s *Function) ClearInputHeader() {
	s.Input.ClearHeaders()
}

func (s *Function) RemoveInputHeader(name string) {
	s.Input.ClearHeaders()
}

func (s *Function) AddInputQuery(required bool, name, note, defaultValue string, optionValues ...string) {
	query := s.GetInputQuery(name)
	if query != nil {
		query.Required = required
		query.Note = note
		query.DefaultValue = defaultValue
		query.Values = optionValues
	} else {
		s.Input.Queries = append(s.Input.Queries, &Query{
			Name:         name,
			Note:         note,
			Required:     required,
			Values:       optionValues,
			DefaultValue: defaultValue,
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

func (s *Function) SetInputExample(v interface{}) {
	s.Input.Example = v
	arg := modelArgument.FromExample(v)
	if arg != nil {
		s.Input.Model = arg.ToModel()
	} else {
		s.Input.Model = make([]*Type, 0)
	}
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
	s.AddOutputError(gtype.ErrException)
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
