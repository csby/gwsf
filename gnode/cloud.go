package gnode

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Cloud interface {
	Connect() error
}

func NewCloud(log gtype.Log, cfg *gcfg.Config) Cloud {
	instance := &innerCloud{}
	instance.SetLog(log)
	instance.cfg = cfg

	instance.nodeChannels = gtype.NewSocketChannelCollection()
	instance.nodeChannel = instance.nodeChannels.NewChannel(nil)
	instance.init()

	return instance
}

type innerCloud struct {
	gtype.Base

	cfg          *gcfg.Config
	isConnected  bool
	dialer       *websocket.Dialer
	nodeChannels gtype.SocketChannelCollection
	nodeChannel  gtype.SocketChannel
}

func (s *innerCloud) init() {
	if s.cfg == nil {
		return
	}

	crt := &Certificate{}
	crt.SetLog(s.GetLog())
	crt.Load(&s.cfg.Node.Cert)

	s.dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  crt.ClientTlsConfig(),
	}
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
		Scheme: "wss",
		Host:   fmt.Sprintf("%s:%d", s.cfg.Node.CloudServer.Address, s.cfg.Node.CloudServer.Port),
		Path:   "/cloud.api/node/connect"}
	u.Query().Add("instance", s.cfg.Node.InstanceId)

	go s.doConnect(u.String())

	return nil
}

func (s *innerCloud) doConnect(url string) {
	for {
		time.Sleep(time.Second)

		s.connect(url)
	}
}

func (s *innerCloud) connect(url string) {
	websocketConn, _, err := s.dialer.Dial(url, nil)
	if err != nil {
		s.LogError("node connect to server fail:", err)
		return
	}
	defer websocketConn.Close()

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("node connect send message to server error:", err)
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
					s.LogError("node connect send message to server error:", err)
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
				s.LogError("node connect read message from server error:", err)
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
					s.LogError("node connect read message from server error:", err)
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}

				if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
					msg := &gtype.SocketMessage{}
					err := json.Unmarshal(msgContent, msg)
					if err != nil {
						s.LogError("node connect socket unmarshal read message from server error:", err)
					} else {
						s.nodeChannels.Read(msg, ch)
					}
				}
			}
		}
	}(waitGroup, websocketConn, s.nodeChannel)

	waitGroup.Wait()
}

func (s *innerCloud) setConnected(isConnected bool) {
	s.isConnected = isConnected
}
