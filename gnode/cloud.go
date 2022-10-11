package gnode

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gclient"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Cloud interface {
	Connect() error
	IsConnected() bool
	SetState(state func(isConnected bool))
	PostJson(uri string, argument interface{}) *gtype.Result
}

func NewCloud(log gtype.Log, cfg *gcfg.Config, dialer *websocket.Dialer, chs *Channels) Cloud {
	instance := &innerCloud{
		isConnected: false,
	}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.chs = chs
	instance.nodeChannel = instance.chs.node.NewChannel(&gtype.Token{ID: cfg.Node.InstanceId})
	instance.dialer = dialer

	return instance
}

type innerCloud struct {
	gtype.Base

	cfg         *gcfg.Config
	isConnected bool
	dialer      *websocket.Dialer
	chs         *Channels
	nodeChannel gtype.SocketChannel

	state func(isConnected bool)
}

func (s *innerCloud) IsConnected() bool {
	return s.isConnected
}

func (s *innerCloud) SetState(state func(isConnected bool)) {
	s.state = state
}

func (s *innerCloud) Connect() error {
	if s.cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if !s.cfg.Node.Enabled {
		return fmt.Errorf("node is disabled")
	}
	if len(s.cfg.Node.CloudServer.Address) < 1 {
		return fmt.Errorf("address of cloud server is empty")
	}
	if s.dialer == nil {
		return fmt.Errorf("websocket dialer is nil")
	}

	u := url.URL{
		Scheme:   "wss",
		Host:     fmt.Sprintf("%s:%d", s.cfg.Node.CloudServer.Address, s.cfg.Node.CloudServer.Port),
		Path:     "/cloud.api/node/connect",
		RawQuery: fmt.Sprintf("instance=%s&version=%s", s.cfg.Node.InstanceId, s.cfg.Module.Version),
	}

	go s.doConnect(u.String(), u.Host)

	return nil
}

func (s *innerCloud) doConnect(uri, host string) {
	for {
		time.Sleep(time.Second)

		s.connect(uri, host)
	}
}

func (s *innerCloud) connect(uri, host string) {
	websocketConn, _, err := s.dialer.Dial(uri, nil)
	if err != nil {
		s.LogDebug("node connect to cloud fail:", err)
		return
	}
	defer func(conn io.Closer, h string) {
		s.setConnected(false)
		conn.Close()
		s.LogInfo("node has been disconnected from cloud: ", h)
	}(websocketConn, host)
	s.LogInfo("node has been connected to cloud: ", host)
	s.setConnected(true)

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
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

				err := conn.WriteJSON(msg)
				if err != nil {
					return
				}
			}
		}
	}(waitGroup, websocketConn, s.nodeChannel)

	// read message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
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
					err := json.Unmarshal(msgContent, msg)
					if err == nil {
						s.chs.node.Read(msg, ch)
					}
				}
			}
		}
	}(waitGroup, websocketConn, s.nodeChannel)

	waitGroup.Wait()
}

func (s *innerCloud) setConnected(isConnected bool) {
	if s.isConnected == isConnected {
		return
	}
	s.isConnected = isConnected

	if s.state != nil {
		go func() {
			defer func() {
				if err := recover(); err != nil {
				}
			}()
			s.state(s.isConnected)
		}()
	}
}

func (s *innerCloud) PostJson(uri string, argument interface{}) *gtype.Result {
	u := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", s.cfg.Node.CloudServer.Address, s.cfg.Node.CloudServer.Port),
		Path:   fmt.Sprintf("/cloud.api%s", uri),
	}

	c := &gclient.Http{}
	if s.dialer != nil {
		c.Transport = &http.Transport{
			TLSClientConfig: s.dialer.TLSClientConfig,
		}
	}

	r := &gtype.Result{}
	_, output, _, code, err := c.PostJson(u.String(), argument)
	if code != 200 {
		r.Code = gtype.ErrInternal.Code()
		r.Error.Summary = gtype.ErrInternal.Summary()
		r.Error.Detail = fmt.Sprintf("cloud response code = %d for '%s' request", code, u.Path)
	} else if err != nil {
		r.Code = gtype.ErrInternal.Code()
		r.Error.Summary = gtype.ErrInternal.Summary()
		r.Error.Detail = err.Error()
	} else {
		err = r.Unmarshal(output)
		if err != nil {
			r.Code = gtype.ErrInternal.Code()
			r.Error.Summary = gtype.ErrInternal.Summary()
			r.Error.Detail = err.Error()
		}
	}

	return r
}
