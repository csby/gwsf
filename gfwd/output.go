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

type Output struct {
	gtype.Base

	Remote gcfg.Cloud
	Target gtype.Forward

	Dialer *websocket.Dialer
}

func (s *Output) Start() {
	go s.run()
}

func (s *Output) forwardConnect(conn net.Conn) {
	defer conn.Close()

	websocketAddr := s.socketUrl()
	websocketConn, _, err := s.Dialer.Dial(websocketAddr, nil)
	if err != nil {
		s.LogError("fwd output connect to server fail:", err)
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
	go func(wg *sync.WaitGroup, dst net.Conn, src *SocketConnection) {
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
	go func(wg *sync.WaitGroup, src net.Conn, dst *SocketConnection) {
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

func (s *Output) run() {
	defer func() {
		if err := recover(); err != nil {
			s.LogError("fwd output error: ", err)
		}
	}()

	addr := fmt.Sprintf("%s:%s", s.Target.TargetAddress, s.Target.TargetPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		s.LogError("fwd output fail: ", err)
		return
	}

	go s.forwardConnect(conn)
}

func (s *Output) socketUrl() string {
	u := url.URL{
		Scheme:   "wss",
		Host:     fmt.Sprintf("%s:%d", s.Remote.Address, s.Remote.Port),
		Path:     "/cloud.api/fwd/response",
		RawQuery: fmt.Sprintf("request=%s&response=%s", s.Target.RequestID, s.Target.ResponseID),
	}

	return u.String()
}
