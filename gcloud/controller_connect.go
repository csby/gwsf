package gcloud

import (
	"encoding/json"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

func (s *Controller) NodeConnect(ctx gtype.Context, ps gtype.Params) {
	websocketConn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		s.LogError("node connect socket connect fail:", err)
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	defer websocketConn.Close()

	now := time.Now()
	crt := ctx.Certificate().Client
	token := &gtype.Token{
		ID:          crt.OrganizationalUnit(),
		UserAccount: crt.Organization(),
		UserName:    crt.CommonName(),
		LoginIP:     ctx.RIP(),
		LoginTime:   now,
		ActiveTime:  now,
	}
	channel := s.chs.OnlineNodes().NewChannel(token)
	defer s.chs.onlineNodes.Remove(channel)

	node := &gtype.Node{}
	node.CopyFrom(token)
	s.writeOptSocketMessage(gtype.WSNodeOnline, node)
	defer s.writeOptSocketMessage(gtype.WSNodeOffline, node)

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("node connect socket send message error:", err)
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
					s.LogError("node connect socket write message error:", err)
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
				s.LogError("node connect send message error:", err)
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
					s.LogError("node connect socket read message error:", err)
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}

				if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
					msg := &gtype.SocketMessage{}
					err := json.Unmarshal(msgContent, msg)
					if err != nil {
						s.LogError("node connect socket unmarshal read message error:", err)
					} else {
						s.chs.onlineNodes.Read(msg, ch)
					}
				}
			}
		}
	}(waitGroup, websocketConn, channel)

	waitGroup.Wait()
}

func (s *Controller) NodeConnectDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "Websocket")
	function := catalog.AddFunction(method, uri, "交互推送")
	function.SetNote("接收或发送节点的交互信息，该接口保持阻塞至连接关闭")
	function.SetRemark("该接口需要客户端证书")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 2})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
