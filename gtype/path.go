package gtype

import (
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"strings"
)

type Path struct {
	Prefix             string
	IsShortPath        bool
	DefaultIsWebsocket bool
	DefaultTokenPlace  int
	DefaultTokenUI     func() []TokenUI
	DefaultTokenCreate func(items []TokenAuth, ctx Context) (string, Error)
}

func (s *Path) Uri(path string, params ...interface{}) Uri {
	sb := &strings.Builder{}
	sb.WriteString(s.Prefix)
	sb.WriteString(path)
	root := sb.String()
	param := fmt.Sprint(params...)

	uri := &uriPath{
		rawPath:     fmt.Sprint(root, param),
		shortPath:   fmt.Sprint(s.toShortUrl(root), param),
		isShortPath: s.IsShortPath,
		tokenPlace:  s.DefaultTokenPlace,
		tokenUI:     s.DefaultTokenUI,
		tokenCreate: s.DefaultTokenCreate,
	}

	return uri
}

func (s *Path) toShortUrl(url string) string {
	h := adler32.New()
	_, err := h.Write([]byte(url))
	if err != nil {
		return url
	}

	return fmt.Sprintf("/%s", hex.EncodeToString(h.Sum(nil)))
}

type uriPath struct {
	rawPath     string
	shortPath   string
	isShortPath bool
	isWebsocket bool

	tokenPlace  int
	tokenUI     func() []TokenUI
	tokenCreate func(items []TokenAuth, ctx Context) (string, Error)
}

func (s *uriPath) Path() string {
	if s.isShortPath {
		return s.shortPath
	} else {
		return s.rawPath
	}
}

func (s *uriPath) IsWebsocket() bool {
	return s.isWebsocket
}

func (s *uriPath) SetIsWebsocket(isWebsocket bool) Uri {
	s.isWebsocket = isWebsocket
	return s
}

func (s *uriPath) TokenPlace() int {
	return s.tokenPlace
}

func (s *uriPath) SetTokenPlace(place int) Uri {
	s.tokenPlace = place
	return s
}

func (s *uriPath) TokenUI() func() []TokenUI {
	return s.tokenUI
}

func (s *uriPath) SetTokenUI(ui func() []TokenUI) Uri {
	s.tokenUI = ui
	return s
}

func (s *uriPath) TokenCreate() func(items []TokenAuth, ctx Context) (string, Error) {
	return s.tokenCreate
}

func (s *uriPath) SetTokenCreate(create func(items []TokenAuth, ctx Context) (string, Error)) Uri {
	s.tokenCreate = create
	return s
}
