package gfwd

import "github.com/csby/gwsf/gcfg"

type Input struct {
	forward

	InstanceID   string
	Local        gcfg.Fwd
	StateChanged func(id string, isRunning bool, lastError string)

	isRunning bool
	lastError string
}

func (s *Input) IsRunning() bool {
	return s.isRunning
}

func (s *Input) LastError() string {
	return s.lastError
}

func (s *Input) setIsRunning(isRunning bool) {
	if s.isRunning == isRunning {
		return
	}
	s.isRunning = isRunning

	if s.StateChanged != nil {
		go func() {
			defer func() {
				if err := recover(); err != nil {
				}
			}()
			s.StateChanged(s.Local.ID, s.isRunning, s.lastError)
		}()
	}
}
