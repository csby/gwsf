package gfwd

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/url"
	"sync"
)

type Input struct {
	gtype.Base

	Local  gcfg.Fwd
	Remote gcfg.Cloud
	Dialer *websocket.Dialer

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

func (s *Input) forwardConnect(conn net.Conn) {
	defer conn.Close()

	websocketAddr := s.socketUrl()
	websocketConn, _, err := s.Dialer.Dial(websocketAddr, nil)
	if err != nil {
		s.LogError("fwd input connect to server fail:", err)
		return
	}
	defer websocketConn.Close()

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	scrConn := &SocketConnection{
		Conn: websocketConn,
	}
	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, src net.Conn, dst *SocketConnection) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("fwd input write error:", err)
			}
			stopRead <- true
		}()

		for {
			select {
			case <-stopWrite:
				return
			default:
				_, err := io.Copy(dst, src)
				if err != nil {
					return
				}
			}
		}
	}(waitGroup, conn, scrConn)

	// read message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, dst net.Conn, src *SocketConnection) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("fwd input write error:", err)
			}
			stopWrite <- true
		}()

		for {
			select {
			case <-stopRead:
				return
			default:
				_, err := io.Copy(dst, src)
				if err != nil {
					return
				}
			}
		}
	}(waitGroup, conn, scrConn)

	waitGroup.Wait()
}

func (s *Input) run() {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("fwd listen error: ", err)
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.Local.ListenAddress, s.Local.ListenPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.LogError("fwd listen fail: ", err)
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
			s.LogError("fwd accept fail: ", err)
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
		Scheme:   "wss",
		Host:     fmt.Sprintf("%s:%d", s.Remote.Address, s.Remote.Port),
		Path:     "/cloud.api/fwd/request",
		RawQuery: fmt.Sprintf("request=%s&addr=%s&port=%d", s.Local.TargetNodeID, s.Local.TargetAddress, s.Local.TargetPort),
	}

	return u.String()
}
