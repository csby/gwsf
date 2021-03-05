package gfwd

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"net"
	"net/url"
)

type Input struct {
	forward

	InstanceID string
	Local      gcfg.Fwd
	Remote     gcfg.Cloud

	listener net.Listener
}

func (s *Input) IsRunning() bool {
	if s.listener == nil {
		return false
	}

	return true
}

func (s *Input) Stop() {
	s.close()
}

func (s *Input) Start() {
	go s.run()
}

func (s *Input) forwardConnect(src net.Conn) {
	addr := s.socketUrl()
	dst, _, err := s.Dialer.Dial(addr, nil)
	if err != nil {
		src.Close()
		return
	}
	defer s.goCloseConn(dst)
	defer s.goCloseConn(src)

	ch := make(chan error, 1)
	go s.copyTcpToSocket(ch, dst, src)
	go s.copySocketToTcp(ch, src, dst)
	<-ch
}

func (s *Input) run() {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("fwd input listen error: ", err)
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.Local.ListenAddress, s.Local.ListenPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.LogError("fwd input listen fail: ", err)
		return
	}
	s.listener = l
	defer s.close()

	s.LogInfo(fmt.Sprintf("forward input(id=%s) is ready, router: '%s' => '%s(%s)' => '%s:%d'",
		s.Local.ID,
		addr,
		s.Local.TargetNodeName, s.Local.TargetNodeID,
		s.Local.TargetAddress, s.Local.TargetPort))

	for {
		c, err := l.Accept()
		if err != nil {
			s.LogError("fwd input accept fail: ", err)
			return
		}

		go s.forwardConnect(c)
	}
}

func (s *Input) close() {
	if s.listener == nil {
		return
	}

	s.listener.Close()
	s.listener = nil

	s.LogInfo(fmt.Sprintf("forward input(id=%s) is closed, router: '%s:%d' => '%s(%s)' => '%s:%d'",
		s.Local.ID,
		s.Local.ListenAddress, s.Local.ListenPort,
		s.Local.TargetNodeName, s.Local.TargetNodeID,
		s.Local.TargetAddress, s.Local.TargetPort))
}

func (s *Input) socketUrl() string {
	u := url.URL{
		Scheme: "wss",
		Host:   fmt.Sprintf("%s:%d", s.Remote.Address, s.Remote.Port),
		Path:   "/cloud.api/fwd/request",
		RawQuery: fmt.Sprintf("instance=%s&node=%s&addr=%s&port=%d",
			s.InstanceID,
			s.Local.TargetNodeID, s.Local.TargetAddress, s.Local.TargetPort),
	}

	return u.String()
}
