package gserver

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
)

func (s *server) Status() (gtype.ServerStatus, error) {
	cfg := &service.Config{
		Name: fmt.Sprintf("%s.service", s.serviceName),
	}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return gtype.ServerStatusUnknown, err
	}

	status, err := svc.Status()
	if err != nil {
		return gtype.ServerStatusUnknown, err
	}

	return gtype.ServerStatus(status), nil
}
