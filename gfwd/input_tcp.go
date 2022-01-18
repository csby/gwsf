package gfwd

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"net"
	"net/url"
)

type InputTcp struct {
	Input

	Remote gcfg.Cloud

	listener net.Listener
}

func (s *InputTcp) IsRunning() bool {
	if s.listener == nil {
		return false
	}

	return true
}

func (s *InputTcp) Stop() {
	s.close()
}

func (s *InputTcp) Start() {
	go s.run()
}

func (s *InputTcp) forwardConnect(src net.Conn) {
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

func (s *InputTcp) run() {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("fwd tcp input listen error: ", err)
		}
		s.setIsRunning(false)
	}()

	s.lastError = ""
	addr := fmt.Sprintf("%s:%d", s.Local.ListenAddress, s.Local.ListenPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.lastError = err.Error()
		s.LogError("fwd tcp input listen fail: ", err)
		return
	}
	s.listener = l
	defer func() {
		s.close()
		s.LogInfo(fmt.Sprintf("forward tcp input(id=%s) is closed, router: '%s:%d' => '%s(%s)' => '%s:%d'",
			s.Local.ID,
			s.Local.ListenAddress, s.Local.ListenPort,
			s.Local.TargetNodeName, s.Local.TargetNodeID,
			s.Local.TargetAddress, s.Local.TargetPort))
	}()

	s.LogInfo(fmt.Sprintf("forward tcp input(id=%s) is ready, router: '%s' => '%s(%s)' => '%s:%d'",
		s.Local.ID,
		addr,
		s.Local.TargetNodeName, s.Local.TargetNodeID,
		s.Local.TargetAddress, s.Local.TargetPort))

	s.setIsRunning(true)
	for {
		c, e := l.Accept()
		if e != nil {
			s.LogError("fwd tcp input accept fail: ", e)
			return
		}

		go s.forwardConnect(c)
	}
}

func (s *InputTcp) close() {
	if s.listener == nil {
		return
	}

	s.listener.Close()
	s.listener = nil
}

func (s *InputTcp) socketUrl() string {
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
