package gfwd

import (
	"errors"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"io"
	"net"
)

type forward struct {
	gtype.Base

	Dialer *websocket.Dialer
}

var errInvalidWrite = errors.New("invalid write result")

func (s *forward) goCloseConn(conn io.Closer) {
	if conn == nil {
		return
	}

	go conn.Close()
}

func (s *forward) copyTcpToSocket(ch chan<- error, dst *websocket.Conn, src net.Conn) (written int64, err error) {
	size := 32 * 1024
	buf := make([]byte, size)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			ew := dst.WriteMessage(websocket.BinaryMessage, buf[0:nr])
			if ew != nil {
				err = ew
				break
			}
			written += int64(nr)
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	ch <- err

	return written, err
}

func (s *forward) copySocketToTcp(ch chan<- error, dst net.Conn, src *websocket.Conn) (written int64, err error) {
	for {
		t, d, er := src.ReadMessage()
		if er != nil {
			err = er
			break
		}
		if t == websocket.CloseMessage {
			err = io.ErrShortWrite
			break
		}
		nr := len(d)
		if nr > 0 {
			nw, ew := dst.Write(d)
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
	}

	ch <- err

	return written, err
}
