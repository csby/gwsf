package gdoc

import (
	"encoding/hex"
	"fmt"
	"github.com/csby/gwsf/gtype"
	"hash/adler32"
	"strings"
)

func NewDoc(enable bool) gtype.Doc {
	return &doc{
		enable:    enable,
		catalogs:  make(CatalogCollection, 0),
		functions: make(map[string]*Function),
	}
}

type doc struct {
	enable    bool
	catalogs  CatalogCollection
	functions map[string]*Function

	onFunctionReady func(index int, method, path, name string)
}

func (s *doc) Enable() bool {
	return s.enable
}

func (s *doc) AddCatalog(name string) gtype.Catalog {
	c := len(s.catalogs)
	for i := 0; i < c; i++ {
		item := s.catalogs[i]
		if item.Name == name {
			return item
		}
	}

	item := &Catalog{Name: name}
	item.OnAddFunction(s.onNewFunction)
	item.Children = make(CatalogCollection, 0)

	s.catalogs = append(s.catalogs, item)

	return item
}

func (s *doc) OnFunctionReady(f func(index int, method, path, name string)) {
	s.onFunctionReady = f
}

func (s *doc) Catalogs() interface{} {
	return s.catalogs
}

func (s *doc) Function(id, schema, host string) (interface{}, error) {
	fun, ok := s.functions[id]
	if ok {
		if fun.IsWebsocket {
			if strings.ToLower(schema) == "https" {
				fun.FullPath = fmt.Sprintf("%s://%s%s", "wss", host, fun.Path)
			} else {
				fun.FullPath = fmt.Sprintf("%s://%s%s", "ws", host, fun.Path)
			}
		} else {
			fun.FullPath = fmt.Sprintf("%s://%s%s", schema, host, fun.Path)
		}
		return fun, nil
	} else {
		return nil, fmt.Errorf("id '%s' not exist", id)
	}
}

func (s *doc) TokenUI(id string) (interface{}, error) {
	fun, ok := s.functions[id]
	if !ok {
		return nil, fmt.Errorf("id '%s' not exist", id)
	}

	ui := fun.TokenUI
	if ui == nil {
		return nil, fmt.Errorf("ui function (id = '%s') not implement", id)
	}

	return ui(), nil
}

func (s *doc) TokenCreate(id string, items []gtype.TokenAuth, ctx gtype.Context) (string, gtype.Error) {
	fun, ok := s.functions[id]
	if !ok {
		return "", gtype.ErrInput.New(fmt.Errorf("id '%s' not exist", id))
	}

	create := fun.TokenCreate
	if create == nil {
		return "", gtype.ErrInternal.New(fmt.Errorf("create function (id = '%s') not implement", id))
	}

	return create(items, ctx)
}

func (s *doc) onNewFunction(fun *Function) {
	id := s.generateFunctionId(fun.Method, fun.Path)
	_, ok := s.functions[id]
	if ok {
		panic(fmt.Sprintf("a document handle is already registered for path '%s: %s'", fun.Method, fun.Path))
	}

	fun.ID = id
	s.functions[id] = fun

	if s.onFunctionReady != nil {
		s.onFunctionReady(len(s.functions), fun.Method, fun.Path, fun.Name)
	}
}

func (s *doc) generateFunctionId(method, path string) string {
	h := adler32.New()
	_, err := h.Write([]byte(method + path))
	if err != nil {
		return path
	}

	return hex.EncodeToString(h.Sum(nil))
}
