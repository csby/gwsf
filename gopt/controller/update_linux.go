package controller

import (
	"github.com/csby/gwsf/gtype"
)

func (s *Update) isEnable() bool {
	return false
}

func (s *Update) info() (*gtype.SvcUpdInfo, gtype.Error) {
	return nil, gtype.ErrNotSupport
}

func (s *Update) canRestart() bool {
	return s.isEnable()
}

func (s *Update) restart() gtype.Error {
	return gtype.ErrNotSupport
}

func (s *Update) canUpdate() bool {
	info, err := s.svcMgr.RemoteInfo()
	if err == nil {
		if info.Interactive {
			return false
		}
	}

	return true
}

func (s *Update) update(ctx gtype.Context) gtype.Error {
	return gtype.ErrNotSupport
}

func (s *Update) executeFileName() string {
	return "gwsfupd"
}
