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

func NewWebsocket(log gtype.Log, cfg *gcfg.Config, db gtype.TokenDatabase, chs gtype.SocketChannelCollection) *Websocket {
	instance := &Websocket{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.dbToken = db
	instance.wsChannels = chs
	instance.wsGrader = websocket.Upgrader{CheckOrigin: instance.checkOrigin}
	instance.input = &appendix{}
	instance.output = &appendix{}

	if chs != nil {
		chs.SetListener(nil, instance.onChannelRemoved)
		//chs.AddReader(instance.onChannelRead)
	}

	return instance
}

type Websocket struct {
	controller

	wsGrader websocket.Upgrader
	input    *appendix
	output   *appendix
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
				Token:       ctx.Token(),
				UserAccount: token.UserAccount,
				UserName:    token.UserName,
				LoginIP:     token.LoginIP,
				LoginTime:   gtype.DateTime(time.Now()),
			},
		}, token)
	}
	channel := s.wsChannels.NewChannel(token)
	defer s.wsChannels.Remove(channel)
	s.wsChannels.Write(&gtype.SocketMessage{ID: gtype.WSOptUserOnline}, nil)
	defer s.wsChannels.Write(&gtype.SocketMessage{ID: gtype.WSOptUserOffline}, nil)

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

				conn.WriteJSON(msg)
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
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}

				if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
					msg := &gtype.SocketMessage{}
					err := json.Unmarshal(msgContent, msg)
					if err == nil {
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
	function.SetInputFormat(gtype.ArgsFmtJson)
	function.SetOutputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputFormat(gtype.ArgsFmtJson)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)

	input := function.SetInputAppendix("消息标识")
	s.input.impl = input
	output := function.SetOutputAppendix("消息标识")
	s.output.impl = output

	item := &appendixItem{}
	output.AddItem(item.Set(gtype.WSOptUserLogin, &gtype.OnlineUser{LoginTime: gtype.DateTime(time.Now())}))
	output.AddItem(item.Set(gtype.WSOptUserLogout, gtype.NewGuid()))
	output.AddItem(item.Set(gtype.WSOptUserOnline, nil))
	output.AddItem(item.Set(gtype.WSOptUserOffline, nil))
	output.AddItem(item.Set(gtype.WSSiteUpload, &gtype.WebApp{}))
	output.AddItem(item.Set(gtype.WSRootSiteUploadFile, &gtype.SiteFile{UploadTime: gtype.DateTime(time.Now())}))
	output.AddItem(item.Set(gtype.WSRootSiteDeleteFile, &gtype.SiteFileFilter{}))
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
	if message == nil || channel == nil {
		return
	}

	channel.Container().Write(&gtype.SocketMessage{
		ID:   message.ID,
		Data: message.Data,
	}, channel.Token())
}
