package gserver

import "net/http"

type notFound struct {
	root string
}

func (s *notFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(s.root) > 0 && r.Method == "GET" {
		http.FileServer(http.Dir(s.root)).ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}
