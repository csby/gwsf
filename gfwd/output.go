package gfwd

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"net"
	"net/url"
)

type Output struct {
	forward

	Remote gcfg.Cloud
	Target gtype.ForwardRequest
}

func (s *Output) Start() {
	go s.run()
}

func (s *Output) forwardConnect(dst net.Conn) {
	addr := s.socketUrl()
	src, _, err := s.Dialer.Dial(addr, nil)
	if err != nil {
		dst.Close()
		s.LogError("fwd output connect to cloud fail:", err)
		return
	}
	defer s.goCloseConn(dst)
	defer s.goCloseConn(src)

	ch := make(chan error, 1)
	go s.copySocketToTcp(ch, dst, src)
	go s.copyTcpToSocket(ch, src, dst)
	err = <-ch
}

func (s *Output) run() {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("fwd output error: ", err)
		}
	}()

	addr := fmt.Sprintf("%s:%s", s.Target.TargetAddress, s.Target.TargetPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		s.LogError("fwd output connect to target fail: ", err)
		return
	}

	go s.forwardConnect(conn)
}

func (s *Output) socketUrl() string {
	u := url.URL{
		Scheme:   "wss",
		Host:     fmt.Sprintf("%s:%d", s.Remote.Address, s.Remote.Port),
		Path:     "/cloud.api/fwd/response",
		RawQuery: fmt.Sprintf("id=%s", s.Target.ID),
	}

	return u.String()
}
