package gproxy

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/csby/tcpproxy"
	"net"
	"time"
)

type Server struct {
	gtype.Base

	Routes                   []Route
	StatusChanged            func(status Status)
	OnConnected              func(link Link)
	OnDisconnected           func(link Link)
	OnTargetAliveChanged     func(item *TargetAddressItem)
	OnTargetConnCountChanged func(item *TargetAddressItem, increase bool)

	agent           *tcpproxy.Proxy
	status          Status
	err             interface{}
	startTime       gtype.DateTime
	targetAddresses []*TargetAddress
	isAliveChecking bool
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

	s.targetAddresses = make([]*TargetAddress, 0)
	routes := s.Routes
	count := len(routes)
	if count < 1 {
		return fmt.Errorf("routes is empty")
	}

	s.agent = &tcpproxy.Proxy{}
	for index := 0; index < count; index++ {
		route := routes[index]
		if len(route.Target) < 1 {
			continue
		}
		address := &TargetAddress{
			SourceId:     route.SourceId,
			TargetId:     route.TargetId,
			AliveChanged: s.OnTargetAliveChanged,
			CountChanged: s.OnTargetConnCountChanged,
		}
		address.SetAddress(route.Target)
		address.AddAddress(route.SpareTargets)
		s.targetAddresses = append(s.targetAddresses, address)

		dest := &TargetProxy{
			Address:              *address,
			ProxyProtocolVersion: route.Version,
			OnConnected:          s.onConnected,
			OnDisconnected:       s.onDisconnected,
		}

		path := ""
		if len(route.Domain) > 0 {
			if route.IsTls {
				s.agent.AddSNIRoute(route.Address, route.Domain, dest)
			} else {
				s.agent.AddHTTPHostRoute(route.Address, route.Domain, route.Path, dest)
				path = route.Path
			}
		} else {
			s.agent.AddRoute(route.Address, dest)
		}

		s.LogInfo(fmt.Sprintf("proxy(version=%d, tls=%v): %s%s, %s => %s",
			route.Version, route.IsTls, route.Domain, path, route.Address, route.Targets()))
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

	if s.isAliveChecking == false {
		s.isAliveChecking = true
		go s.doAliveChecking()
	}

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
	s.targetAddresses = make([]*TargetAddress, 0)
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
			Status:     0,
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
			Status:     1,
		})
	}
}

func (s *Server) onTargetConnCountChanged(item *TargetAddressItem, increase bool) {
	if s.OnTargetConnCountChanged != nil {
		s.OnTargetConnCountChanged(item, increase)
	}
}

func (s *Server) onTargetAliveChanged(item *TargetAddressItem) {
	if s.OnTargetAliveChanged != nil {
		s.OnTargetAliveChanged(item)
	}
}

func (s *Server) isAlive(addr string) error {
	if len(addr) < 1 {
		return fmt.Errorf("address is empty")
	}

	timeout := 500 * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func (s *Server) doAliveChecking() {
	defer func() {
		if err := recover(); err != nil {
			s.isAliveChecking = false
		}
	}()

	interval := 5 * time.Second
	for {
		time.Sleep(interval)

		if s.status != StatusRunning {
			continue
		}

		targetAddresses := s.targetAddresses
		tc := len(targetAddresses)
		for ti := 0; ti < tc; ti++ {
			if s.status != StatusRunning {
				break
			}

			targetAddress := targetAddresses[ti]
			if targetAddress == nil {
				continue
			}

			items := targetAddress.Items()
			c := len(items)
			for i := 0; i < c; i++ {
				item := items[i]
				if item == nil {
					continue
				}
				err := s.isAlive(item.Addr)
				if err != nil {
					item.SetAlive(false)
				} else {
					item.SetAlive(true)
				}
			}
		}

	}
}
