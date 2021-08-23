package gserver

import (
	"github.com/csby/gsecurity/gcrt"
	"net/http"
)

type protocol struct {
	handler *handler

	caCrt     *gcrt.Crt
	serverCrt *gcrt.Pfx
}

func (s *protocol) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.handler == nil {
		return
	}

	s.handler.ServeHTTP(w, r, s.caCrt, s.serverCrt)
}
