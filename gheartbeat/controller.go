package gheartbeat

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

func NewController(log gtype.Log, cfg *gcfg.Config, opt gtype.SocketChannelCollection) *Controller {
	inst := &Controller{}
	inst.SetLog(log)
	inst.cfg = cfg
	inst.optWcs = opt
	inst.wsGrader = websocket.Upgrader{CheckOrigin: inst.checkOrigin}
	inst.heartbeats = &gtype.HeartbeatArray{Items: make([]*gtype.Heartbeat, 0)}

	return inst
}

type Controller struct {
	gtype.Base

	cfg *gcfg.Config

	wsGrader   websocket.Upgrader
	heartbeats *gtype.HeartbeatArray
	optWcs     gtype.SocketChannelCollection
}

func (s *Controller) Connect(ctx gtype.Context, ps gtype.Params) {
	websocketConn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	defer websocketConn.Close()

	heartbeat := &gtype.Heartbeat{
		Id:       gtype.NewGuid(),
		Protocol: ctx.Schema(),
		Host:     ctx.Host(),
		Remote:   ctx.RIP(),
	}
	s.heartbeats.Add(heartbeat)
	go s.writeOptSocketMessage(gtype.WSHeartbeatConnected, heartbeat)
	defer func(id string) {
		s.heartbeats.Del(id)
		go s.writeOptSocketMessage(gtype.WSHeartbeatDisconnected, id)
	}(heartbeat.Id)

	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn) {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
			}
		}()

		for {
			msgType, msgContent, re := conn.ReadMessage()
			if re != nil {
				return
			}
			if msgType == websocket.CloseMessage {
				return
			}

			if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
				conn.WriteMessage(msgType, msgContent)
			}
		}
	}(waitGroup, websocketConn)

	waitGroup.Wait()
}

func (s *Controller) ConnectDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc)
	function := catalog.AddFunction(method, uri, "连接")
	function.SetNote("将接收到的信息发送给发送方，该接口保持阻塞至连接关闭")
	function.SetRemark("一般用于负载均衡器进行主机服务状态检测")
	function.SetInputExample(&gtype.SocketMessage{ID: 0})
	function.SetInputFormat(gtype.ArgsFmtJson)
	function.SetOutputExample(&gtype.SocketMessage{ID: 0})
	function.SetOutputFormat(gtype.ArgsFmtJson)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *Controller) createCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog(ApiCatalog)

	count := len(names)
	if count < 1 {
		return root
	}

	child := root
	for i := 0; i < count; i++ {
		name := names[i]
		child = child.AddChild(name)
	}

	return child
}

func (s *Controller) createOptCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
	root := doc.AddCatalog("管理平台接口")

	count := len(names)
	if count < 1 {
		return root
	}

	child := root
	for i := 0; i < count; i++ {
		name := names[i]
		child = child.AddChild(name)
	}

	return child
}

func (s *Controller) checkOrigin(r *http.Request) bool {
	if r != nil {
	}

	return true
}

func (s *Controller) writeOptSocketMessage(id int, data interface{}) bool {
	if s.optWcs == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.optWcs.Write(msg, nil)

	return true
}
