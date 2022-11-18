package gcluster

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"io"
	"sync"
	"time"
)

type Connection struct {
	gtype.Base

	Index   uint64
	Channel gtype.SocketChannel
	Readers gtype.SocketChannelCollection

	Url    string
	Host   string
	Dialer *websocket.Dialer

	State func(conn *Connection)

	connected bool
	version   string
}

func (s *Connection) Connected() bool {
	return s.connected
}

func (s *Connection) Version() string {
	return s.version
}

func (s *Connection) OnConnect(conn *websocket.Conn, version string) {
	if conn == nil {
		return
	}
	if s.Channel == nil {
		return
	}
	if s.Readers == nil {
		return
	}
	if s.connected {
		return
	}
	s.version = version

	defer s.setConnected(false)
	s.setConnected(true)

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
				s.LogError("cluster socket send message error:", e)
			}
			stopRead <- true
		}()

		for {
			select {
			case <-stopWrite:
				return
			case msg, ok := <-ch.Read():
				if !ok {
					return
				}

				conn.WriteJSON(msg)
			}
		}
	}(waitGroup, conn, s.Channel)

	// read message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
				s.LogError("cluster socket read message error:", e)
			}
			stopWrite <- true
		}()

		for {
			select {
			case <-stopRead:
				return
			default:
				msgType, msgContent, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}

				if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
					msg := &gtype.SocketMessage{}
					err = json.Unmarshal(msgContent, msg)
					if err == nil {
						s.Readers.Read(msg, ch)
					}
				}
			}
		}
	}(waitGroup, conn, s.Channel)

	waitGroup.Wait()
}

func (s *Connection) DoConnect() error {
	if s.Channel == nil {
		return fmt.Errorf("channel is nil")
	}
	if s.Dialer == nil {
		return fmt.Errorf("dialer is nil")
	}
	if len(s.Url) < 1 {
		return fmt.Errorf("url is empty")
	}

	go s.doConnect(s.Url, s.Host)

	return nil
}

func (s *Connection) doConnect(uri, host string) {
	for {
		time.Sleep(time.Second)

		s.connect(uri, host)
	}
}

func (s *Connection) connect(url, host string) {
	websocketConn, _, err := s.Dialer.Dial(url, nil)
	if err != nil {
		s.LogDebug("instance connect to cluster fail:", err)
		return
	}
	defer func(conn io.Closer, h string) {
		s.setConnected(false)
		conn.Close()
		s.LogInfo("instance has been disconnected from cluster: ", h)
	}(websocketConn, host)
	s.LogInfo("instance has been connected to cluster: ", host)
	s.setConnected(true)

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
			}
			stopRead <- true
		}()

		for {
			select {
			case <-stopWrite:
				return
			case msg, ok := <-ch.Read():
				if !ok {
					return
				}

				we := conn.WriteJSON(msg)
				if we != nil {
					return
				}
			}
		}
	}(waitGroup, websocketConn, s.Channel)

	// read message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
			}
			stopWrite <- true
		}()

		for {
			select {
			case <-stopRead:
				return
			default:
				msgType, msgContent, re := conn.ReadMessage()
				if re != nil {
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}

				if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
					msg := &gtype.SocketMessage{}
					re = json.Unmarshal(msgContent, msg)
					if re == nil {
						if s.Readers != nil {
							s.Readers.Read(msg, ch)
						}
					}
				}
			}
		}
	}(waitGroup, websocketConn, s.Channel)

	waitGroup.Wait()
}

func (s *Connection) setConnected(connected bool) {
	if s.connected == connected {
		return
	}
	s.connected = connected

	if s.State != nil {
		go func() {
			defer func() {
				if err := recover(); err != nil {
				}
			}()
			s.State(s)
		}()
	}
}
