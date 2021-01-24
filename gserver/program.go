package gserver

import (
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
)

type program struct {
	gtype.Base

	host *host
}

func (s *program) Start(svc service.Service) error {
	s.LogInfo("service '", svc.String(), "' started")

	go s.run()

	return nil
}

func (s *program) Stop(svc service.Service) error {
	s.LogInfo("service '", svc.String(), "' stopped")

	return nil
}

func (s *program) run() {
	if s.host == nil {
		return
	}

	err := s.host.Run()
	if err != nil {
		s.LogError(err)
	}
}
