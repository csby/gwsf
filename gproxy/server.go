package gproxy

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/csby/tcpproxy"
	"time"
)

type Server struct {
	gtype.Base

	Routes         []Route
	StatusChanged  func(status Status)
	OnConnected    func(link Link)
	OnDisconnected func(link Link)

	agent     *tcpproxy.Proxy
	status    Status
	err       interface{}
	startTime gtype.DateTime
}

func (s *Server) Start() error {
	defer func() {
		if err := recover(); err != nil {
			s.setStatus(StatusStopped)
			s.err = err
		}
	}()

	if s.status != StatusStopped {
		return fmt.Errorf("server is %s", s.status)
	}

	routes := s.Routes
	count := len(routes)
	if count < 1 {
		return fmt.Errorf("routes is empty")
	}

	s.agent = &tcpproxy.Proxy{}
	for index := 0; index < count; index++ {
		route := routes[index]
		dest := &tcpproxy.DialProxy{
			Addr:                 route.Target,
			ProxyProtocolVersion: route.Version,
			OnConnected:          s.onConnected,
			OnDisconnected:       s.onDisconnected,
		}

		if len(route.Domain) > 0 {
			if route.IsTls {
				s.agent.AddSNIRoute(route.Address, route.Domain, dest)
			} else {
				s.agent.AddHTTPHostRoute(route.Address, route.Domain, dest)
			}
		} else {
			s.agent.AddRoute(route.Address, dest)
		}

		s.LogInfo(fmt.Sprintf("proxy(version=%d, tls=%v): %s, %s => %s",
			route.Version, route.IsTls, route.Domain, route.Address, route.Target))
	}

	s.setStatus(StatusStarting)
	err := s.agent.Start()
	if err != nil {
		s.err = err
		s.setStatus(StatusStopped)
		s.LogError("start proxy server error:", err)
		return err
	}
	s.err = nil
	s.startTime = gtype.DateTime(time.Now())
	s.setStatus(StatusRunning)

	go func() {
		s.agent.Wait()
		s.setStatus(StatusStopped)
	}()

	return nil
}

func (s *Server) Stop() error {
	if s.status != StatusRunning {
		return fmt.Errorf("server has be %s", s.status)
	}

	s.setStatus(StatusStopping)
	return s.agent.Close()
}

func (s *Server) Restart() error {
	if s.status == StatusRunning {
		s.setStatus(StatusStopping)
		s.agent.Close()
	}

	for s.status != StatusStopped {
		time.Sleep(100)
	}

	return s.Start()
}

func (s *Server) Result() *Result {
	result := &Result{Status: s.status}
	if s.status == StatusRunning {
		result.StartTime = &s.startTime
	}
	if s.status == StatusStopped {
		if s.err != nil {
			result.Error = fmt.Sprint(s.err)
		}
	}

	return result
}

func (s *Server) setStatus(status Status) {
	if s.status == status {
		return
	}
	s.status = status

	if s.StatusChanged != nil {
		s.StatusChanged(status)
	}
}

func (s *Server) onConnected(id, addr, domain, source, target string) {
	if s.OnConnected != nil {
		s.OnConnected(Link{
			Id:         id,
			Time:       gtype.DateTime(time.Now()),
			ListenAddr: addr,
			Domain:     domain,
			SourceAddr: source,
			TargetAddr: target,
		})
	}
}

func (s *Server) onDisconnected(id, addr, domain, source, target string) {
	if s.OnDisconnected != nil {
		s.OnDisconnected(Link{
			Id:         id,
			Time:       gtype.DateTime(time.Now()),
			ListenAddr: addr,
			Domain:     domain,
			SourceAddr: source,
			TargetAddr: target,
		})
	}
}
