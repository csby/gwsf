package gserver

import (
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
	"strings"
)

func (s *server) Status() (gtype.ServerStatus, error) {
	name := s.serviceName
	cfg := &service.Config{
		//Name: fmt.Sprintf("%s.service", s.serviceName),
		Name: name,
	}
	svc, err := service.New(nil, cfg)
	if err != nil {
		return gtype.ServerStatusUnknown, err
	}

	status, err := svc.Status()
	if err != nil {
		if err == service.ErrNotInstalled || err.Error() == "the service is not installed" {
			if strings.Contains(name, "@") {
				return gtype.ServerStatusStopped, nil
			}
			return gtype.ServerStatusUnknown, nil
		} else if err.Error() == "service in failed state" {
			return gtype.ServerStatusStopped, nil
		} else if err.Error() == "the service is not installed" {
			return gtype.ServerStatusStopped, nil
		} else if strings.Contains(name, "@") {
			return gtype.ServerStatusStopped, nil
		}
		return gtype.ServerStatusUnknown, err
	}

	return gtype.ServerStatus(status), nil
}
