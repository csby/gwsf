package controller

import (
	"encoding/json"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gcloud"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gnode"
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
		chs.AddReader(instance.onChannelRead)
		chs.AddFilter(instance.onChannelFilter)
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

	ctx.FireAfterInput()

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if re := recover(); re != nil {
				s.LogError("notify subscribe socket send message error:", re)
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
			if re := recover(); re != nil {
				s.LogError("notify subscribe socket send message error:", re)
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

	s.appendInput(input)
	s.appendOutput(output)
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

	switch message.ID {
	case gtype.WSNetworkThroughput, gtype.WSCpuUsage, gtype.WSMemUsage:
		{
			data := false
			err := message.GetData(&data)
			if err != nil {
				break
			}
			channel.Subscribe(message.ID, data)
		}
		break
	}
}

func (s *Websocket) onChannelFilter(message *gtype.SocketMessage, channel gtype.SocketChannel, token *gtype.Token) bool {
	if message == nil || channel == nil {
		return false
	}

	switch message.ID {
	case gtype.WSNetworkThroughput, gtype.WSCpuUsage, gtype.WSMemUsage:
		count := channel.Subscription(message.ID)
		if count < 1 {
			return true
		}
	}

	return false
}

func (s *Websocket) appendInput(v gtype.Appendix) {
	if v == nil {
		return
	}

	v.Add(gtype.WSNetworkThroughput, "WSNetworkThroughput", "网络吞吐量(是否订阅)", &gmodel.NotifySubscription{ID: gtype.WSNetworkThroughput, Data: true})
	v.Add(gtype.WSCpuUsage, "WSCpuUsage", "CPU使用率(是否订阅)", &gmodel.NotifySubscription{ID: gtype.WSCpuUsage, Data: true})
	v.Add(gtype.WSMemUsage, "WSMemUsage", "内存使用率(是否订阅)", &gmodel.NotifySubscription{ID: gtype.WSMemUsage, Data: true})
}

func (s *Websocket) appendOutput(v gtype.Appendix) {
	if v == nil {
		return
	}

	item := &appendixItem{}
	v.AddItem(item.Set(gtype.WSClusterNodeStatusChanged, &gtype.ClusterNodeStatus{}))

	v.AddItem(item.Set(gtype.WSHeartbeatConnected, &gtype.Heartbeat{}))
	v.AddItem(item.Set(gtype.WSHeartbeatDisconnected, gtype.NewGuid()))

	v.AddItem(item.Set(gtype.WSOptUserLogin, &gtype.OnlineUser{LoginTime: gtype.DateTime(time.Now())}))
	v.AddItem(item.Set(gtype.WSOptUserLogout, gtype.NewGuid()))
	v.AddItem(item.Set(gtype.WSOptUserOnline, nil))
	v.AddItem(item.Set(gtype.WSOptUserOffline, nil))
	v.AddItem(item.Set(gtype.WSSiteUpload, &gtype.WebApp{}))
	v.AddItem(item.Set(gtype.WSRootSiteUploadFile, &gtype.SiteFile{UploadTime: gtype.DateTime(time.Now())}))
	v.AddItem(item.Set(gtype.WSRootSiteDeleteFile, &gtype.SiteFileFilter{}))

	v.AddItem(item.Set(gtype.WSNodeForwardTcpStart, &gtype.ForwardInfo{}))
	v.AddItem(item.Set(gtype.WSNodeForwardTcpEnd, &gtype.ForwardId{}))

	v.AddItem(item.Set(gtype.WSNodeRegister, &gcloud.Node{}))
	v.AddItem(item.Set(gtype.WSNodeRevoke, &gcloud.NodeDelete{}))
	v.AddItem(item.Set(gtype.WSNodeModify, &gcloud.NodeModify{}))
	v.AddItem(item.Set(gtype.WSNodeInstanceOnline, &gcloud.NodeInstance{}))
	v.AddItem(item.Set(gtype.WSNodeInstanceOffline, &gcloud.NodeInstance{}))

	v.AddItem(item.Set(gtype.WSNodeOnlineStateChanged, &gnode.NodeOnlineState{}))
	v.AddItem(item.Set(gtype.WSNodeFwdInputListenSvcState, &gtype.ForwardState{}))

	v.AddItem(item.Set(gtype.WSNetworkThroughput, &gmodel.MonitorNetworkIOThroughputArgument{}))
	v.AddItem(item.Set(gtype.WSCpuUsage, &gmodel.MonitorCpuPercent{}))
	v.AddItem(item.Set(gtype.WSMemUsage, &gmodel.MonitorMemoryPercent{}))
}
