package gserver

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/kardianos/service"
)

func NewServer(log gtype.Log, cfg *gcfg.Config, handler gtype.Handler) (gtype.Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("invalid parameter: cfg is nil")
	}

	instance := &server{serviceName: cfg.Svc.Name}
	instance.SetLog(log)

	instance.program.SetLog(log)
	instance.program.host = &host{cfg: cfg, httpHandler: handler}
	instance.program.host.SetLog(log)

	svcCfg := &service.Config{
		Name:        instance.serviceName,
		DisplayName: instance.serviceName,
		Description: cfg.Module.Remark,
	}
	svc, err := service.New(&instance.program, svcCfg)
	if err != nil {
		return nil, err
	}
	instance.service = svc

	return instance, nil
}

type server struct {
	gtype.Base

	program     program
	service     service.Service
	serviceName string
}

func (s *server) ServiceName() string {
	return s.serviceName
}

func (s *server) Interactive() bool {
	return service.Interactive()
}

func (s *server) Run() error {
	if s.program.host == nil {
		return fmt.Errorf("invalid host: nil")
	}

	if s.Interactive() {
		return s.program.host.Run()
	} else {
		return s.service.Run()
	}
}

func (s *server) Shutdown() error {
	if s.program.host == nil {
		return fmt.Errorf("invalid host: nil")
	}

	return s.program.host.Close()
}

func (s *server) Restart() error {
	return s.service.Restart()
}

func (s *server) Start() error {
	return s.service.Start()
}

func (s *server) Stop() error {
	return s.service.Stop()
}

func (s *server) Install() error {
	return s.service.Install()
}

func (s *server) Uninstall() error {
	err := s.service.Stop()
	if err != nil {
	}

	return s.service.Uninstall()
}
