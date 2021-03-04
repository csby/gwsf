package gfwd

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
)

type SocketConnection struct {
	*websocket.Conn
}

func (s *SocketConnection) Write(p []byte) (int, error) {
	err := s.Conn.WriteMessage(websocket.BinaryMessage, p)

	return len(p), err
}

func (s *SocketConnection) Read(p []byte) (int, error) {
	t, d, e := s.Conn.ReadMessage()
	if e != nil {
		return len(d), e
	}
	if t == websocket.CloseMessage {
		return len(d), fmt.Errorf("closed")
	} else if t == websocket.BinaryMessage {
		buf := &bytes.Buffer{}
		buf.Write(d)

		return buf.Read(p)
	}

	return 0, nil
}
