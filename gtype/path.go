package gtype

import (
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"strings"
)

type Path struct {
	Prefix      string
	IsShortPath bool
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
}

func (s *uriPath) Path() string {
	if s.isShortPath {
		return s.shortPath
	} else {
		return s.rawPath
	}
}
