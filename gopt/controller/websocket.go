package controller

import (
	"encoding/json"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Websocket struct {
	controller

	wsGrader websocket.Upgrader
}

func NewWebsocket(log gtype.Log, cfg *gcfg.Config, db gtype.TokenDatabase, chs gtype.SocketChannelCollection) *Websocket {
	instance := &Websocket{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dbToken = db
	instance.wsChannels = chs
	instance.wsGrader = websocket.Upgrader{CheckOrigin: instance.checkOrigin}

	if chs != nil {
		chs.SetListener(nil, instance.onChannelRemoved)
		//chs.AddReader(instance.onChannelRead)
	}

	return instance
}

func (s *Websocket) Notify(ctx gtype.Context, ps gtype.Params) {
	websocketConn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		s.LogError("notify subscribe socket connect fail:", err)
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	defer websocketConn.Close()

	token := s.getToken(ctx.Token())
	if token != nil {
		s.dbToken.Permanent(token.ID, true)

		s.wsChannels.Write(&gtype.SocketMessage{
			ID: gtype.WSOptUserLogin,
			Data: &gtype.OnlineUser{
				UserAccount: token.UserAccount,
				UserName:    token.UserName,
				LoginIP:     token.LoginIP,
				LoginTime:   gtype.DateTime(time.Now()),
			},
		}, token)
	}
	channel := s.wsChannels.NewChannel(token)
	defer s.wsChannels.Remove(channel)

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("notify subscribe socket send message error:", err)
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
					s.LogError("notify subscribe socket write message error:", err)
				}
			}
		}
	}(waitGroup, websocketConn, channel)

	// read message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("notify subscribe socket send message error:", err)
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
					s.LogError("notify subscribe socket read message error:", err)
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}

				if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
					msg := &gtype.SocketMessage{}
					err := json.Unmarshal(msgContent, msg)
					if err != nil {
						s.LogError("notify subscribe socket unmarshal read message error:", err)
					} else {
						s.wsChannels.Read(msg, ch)
					}
				}
			}
		}
	}(waitGroup, websocketConn, channel)

	waitGroup.Wait()
}

func (s *Websocket) NotifyDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "Websocket")
	function := catalog.AddFunction(method, uri, "通知推送")
	function.SetNote("订阅并接收系统推送的通知，该接口保持阻塞至连接关闭")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 1})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Websocket) checkOrigin(r *http.Request) bool {
	if r != nil {
	}
	return true
}

func (s *Websocket) onChannelRemoved(channel gtype.SocketChannel) {
	if channel == nil {
		return
	}

	token := channel.Token()
	if token == nil {
		return
	}

	if token.Usage > 0 {
		return
	}

	if s.dbToken != nil {
		s.dbToken.Permanent(token.ID, false)
	}
}

func (s *Websocket) onChannelRead(message *gtype.SocketMessage, channel gtype.SocketChannel) {
	channel.Container().Write(&gtype.SocketMessage{
		ID:   message.ID,
		Data: message.Data,
	}, channel.Token())
}
