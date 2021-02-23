package test

import (
	"github.com/csby/gwsf/gserver"
	"github.com/csby/gwsf/gtype"
)

type Server struct {
}

func (s *Server) Run(starter func(server gtype.Server)) error {
	handler := &Handler{}
	handler.SetLog(log)
	server, err := gserver.NewServer(log, &cfg.Config, handler)
	if err != nil {
		return err
	}

	if starter != nil {
		go starter(server)
	}

	return server.Run()
}
