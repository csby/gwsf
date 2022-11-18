package gcluster

import "github.com/csby/gwsf/gtype"

type Instance struct {
	gtype.Base

	Index uint64
	In    *Connection
	Out   *Connection

	ConnState func(inst *Instance)
}

func (s *Instance) onConnStatusChanged(*Connection) {
	if s.ConnState != nil {
		s.ConnState(s)
	}
}
