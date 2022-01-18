package gcloud

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

func (s *Controller) NodeConnect(ctx gtype.Context, ps gtype.Params) {
	instanceId := ctx.Query("instance")
	if len(instanceId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("instance id of note is empty"))
		return
	}
	version := ctx.Query("version")

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
		ID:          instanceId,
		Version:     version,
		UserAccount: crt.Organization(),
		UserName:    crt.CommonName(),
		LoginIP:     ctx.RIP(),
		LoginTime:   now,
		ActiveTime:  now,
		Ext: gtype.NodeId{
			Instance:    instanceId,
			Certificate: crt.OrganizationalUnit(),
		},
	}
	channel := s.chs.node.NewChannel(token)
	defer s.chs.node.Remove(channel)

	node := &gtype.Node{}
	node.CopyFrom(token)
	node.Province = crt.Province()
	node.Locality = crt.Locality()
	node.Address = crt.StreetAddress()
	notAfter := crt.NotAfter()
	if notAfter != nil {
		v := gtype.DateTime(*notAfter)
		node.CrtNotAfter = &v
	}

	s.clients.NodeOnline(node)
	defer s.clients.NodeOffline(node)

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
	}(waitGroup, websocketConn, channel)

	waitGroup.Wait()
}

func (s *Controller) NodeConnectDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "节点服务")
	function := catalog.AddFunction(method, uri, "节点登录")
	function.SetNote("接收或发送节点的交互信息，该接口保持阻塞至连接关闭")
	function.SetRemark("该接口需要客户端证书")
	function.AddInputQuery(true, "instance", "节点实例ID", "")
	function.AddInputQuery(false, "version", "节点版本号", "")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 2})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
