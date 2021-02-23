package gserver

import (
	"github.com/csby/gwsf/gtype"
)

func (s *server) Status() (gtype.ServerStatus, error) {
	status, err := s.service.Status()
	if err != nil {
		return gtype.ServerStatusUnknown, err
	}

	return gtype.ServerStatus(status), nil
}
