package gdoc

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"sort"
)

const (
	typeCatalog  = 0
	typeFunction = 1
)

type Catalog struct {
	ID       string            `json:"id"`       // 标识
	Name     string            `json:"name"`     // 名称
	Note     string            `json:"note"`     // 说明
	Type     int               `json:"type"`     // 类别: 0-catalog; 1-function
	Keywords string            `json:"keywords"` // 关键字, 用于过滤
	Children CatalogCollection `json:"children"`

	index         int
	onAddFunction func(fun *Function)
}

func (s *Catalog) AddChild(name string) gtype.Catalog {
	c := len(s.Children)
	for i := 0; i < c; i++ {
		item := s.Children[i]
		if item.Name == name {
			return item
		}
	}

	item := &Catalog{Name: name}
	item.Children = make(CatalogCollection, 0)
	item.Type = typeCatalog
	item.Keywords = name
	item.index = len(s.Children)
	item.onAddFunction = s.onAddFunction

	s.Children = append(s.Children, item)
	sort.Sort(s.Children)

	return item
}

func (s *Catalog) AddFunction(method string, uri gtype.Uri, name string) gtype.Function {
	path := uri.Path()
	item := &Catalog{Name: name}
	item.Children = make(CatalogCollection, 0)
	item.Type = typeFunction
	item.Keywords = fmt.Sprintf("%s%s", name, path)
	item.index = len(s.Children)

	s.Children = append(s.Children, item)
	sort.Sort(s.Children)

	fuc := &Function{
		Method:      method,
		Path:        path,
		Name:        name,
		IsWebsocket: uri.IsWebsocket(),
		TokenUI:     uri.TokenUI(),
		TokenCreate: uri.TokenCreate(),
	}
	if fuc.IsWebsocket {
		fuc.Method = "WEBSOCKET"
	}
	fuc.Input = &Input{
		Headers: make([]*Header, 0),
		Queries: make([]*Query, 0),
		Forms:   make([]*Form, 0),
	}
	fuc.Output = &Output{
		Headers: make([]*Header, 0),
		Errors:  make(ErrorCollection, 0),
	}

	if fuc.TokenUI != nil && fuc.TokenCreate != nil {
		if uri.TokenPlace() == gtype.TokenPlaceHeader {
			fuc.AddInputHeader(true, gtype.TokenName, gtype.TokenNote, gtype.TokenValue)
		} else if uri.TokenPlace() == gtype.TokenPlaceQuery {
			fuc.AddInputQuery(true, gtype.TokenName, gtype.TokenNote, gtype.TokenValue)
		}
	}

	//fuc.SetTokenType(httpPath.TokenType())
	//if method == "POST" {
	//	fuc.SetInputContentType(gtype.ContentTypeJson)
	//	fuc.AddOutputHeader("access-control-allow-origin", "*")
	//	fuc.AddOutputHeader(headContentType, "application/json;charset=utf-8")
	//}
	if s.onAddFunction != nil {
		s.onAddFunction(fuc)
		item.ID = fuc.ID
	}

	return fuc
}

func (s *Catalog) OnAddFunction(f func(fun *Function)) {
	s.onAddFunction = f
}

type CatalogCollection []*Catalog

func (s CatalogCollection) Len() int {
	return len(s)
}

func (s CatalogCollection) Less(i, j int) bool {
	if s[i].Type < s[j].Type {
		return true
	} else if s[i].Type == s[j].Type {
		return s[i].index < s[j].index
	}

	return false
}

func (s CatalogCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type CatalogTree struct {
	ID       string `json:"id"`       // 标识
	Name     string `json:"name"`     // 名称
	Note     string `json:"note"`     // 说明
	Type     int    `json:"type"`     // 类别: 0-catalog; 1-function
	Keywords string `json:"keywords"` // 关键字, 用于过滤

	Children []*CatalogTree `json:"children"`
}
