package gcloud

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

func (s *Controller) NodeForwardRequest(ctx gtype.Context, ps gtype.Params) {
	certificateId := ctx.Query("id")
	if len(certificateId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("certificate id of note is empty"))
		return
	}
	targetAddr := ctx.Query("addr")
	if len(targetAddr) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("address of target is empty"))
		return
	}
	targetPort := ctx.Query("port")
	if len(targetPort) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("port of target is empty"))
		return
	}

	targetNode := s.chs.node.OnlineNode(certificateId)
	if targetNode == nil {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("target node(id=%s) not exist", certificateId))
		return
	}

	websocketConn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		s.LogError("node connect socket connect fail:", err)
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	defer websocketConn.Close()

	requestId := gtype.NewGuid()
	responseId := gtype.NewGuid()
	now := time.Now()
	crt := ctx.Certificate().Client
	token := &gtype.Token{
		ID:          requestId,
		UserAccount: crt.Organization(),
		UserName:    crt.CommonName(),
		LoginIP:     ctx.RIP(),
		LoginTime:   now,
		ActiveTime:  now,
		Ext: gtype.NodeId{
			Instance:    requestId,
			Certificate: crt.OrganizationalUnit(),
		},
	}
	requestChannel := s.chs.forward.NewChannel(token)
	defer s.chs.node.Remove(requestChannel)
	responseChannel := s.chs.forward.NewChannel(&gtype.Token{ID: responseId})
	defer s.chs.node.Remove(responseChannel)

	node := &gtype.Node{}
	node.CopyFrom(token)
	s.writeOptSocketMessage(gtype.WSNodeForwardRequest, node)
	defer s.writeOptSocketMessage(gtype.WSNodeForwardRequestEnd, node)

	s.writeNodeSocketMessage(targetNode.ID.Certificate, gtype.WSNodeForwardRequest, &gtype.Forward{
		RequestID:         requestId,
		ResponseID:        responseId,
		NodeCertificateID: targetNode.ID.Certificate,
		NodeInstanceID:    targetNode.ID.Instance,
		TargetAddress:     targetAddr,
		TargetPort:        targetPort,
	})

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// read message (request <= response)
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("forward request write error:", err)
			}
			stopRead <- true
		}()

		for {
			select {
			case <-stopWrite:
				return
			case <-ch.IsStop():
				return
			case data, ok := <-ch.ReadData():
				if !ok {
					return
				}

				err := conn.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					return
				}
			}
		}
	}(waitGroup, websocketConn, requestChannel)

	// write message (request => response)
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn) {
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

				if msgType == websocket.BinaryMessage {
					if !s.chs.forward.WriteData(msgContent, responseId) {
						return
					}
				}
			}
		}
	}(waitGroup, websocketConn)

	waitGroup.Wait()
}

func (s *Controller) NodeForwardRequestDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "节点服务")
	function := catalog.AddFunction(method, uri, "转发请求")
	function.SetNote("接收或发送节点的转发数据，该接口保持阻塞至连接关闭")
	function.SetRemark("该接口需要客户端证书")
	function.AddInputQuery(true, "id", "目标节点证书ID", "")
	function.AddInputQuery(true, "addr", "目标地址", "")
	function.AddInputQuery(true, "port", "目标端口", "")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 2})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Controller) NodeForwardResponse(ctx gtype.Context, ps gtype.Params) {
	requestId := ctx.Query("request")
	if len(requestId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("request id of forward is empty"))
		return
	}
	responseId := ctx.Query("response")
	if len(responseId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("response id of forward is empty"))
		return
	}

	responseChannel := s.chs.forward.Get(responseId)
	if responseChannel == nil {
		s.LogError(fmt.Errorf("response(id=%s) change not exit", responseId))
		ctx.Error(gtype.ErrInternal, fmt.Errorf("response(id=%s) change not exit", responseId))
		return
	}
	if !s.chs.forward.ChannelExist(requestId) {
		s.LogError(fmt.Errorf("request(id=%s) change not exit", requestId))
		ctx.Error(gtype.ErrInternal, fmt.Errorf("request(id=%s) change not exit", requestId))
		return
	}

	websocketConn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		s.LogError("node connect socket connect fail:", err)
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	defer websocketConn.Close()

	waitGroup := &sync.WaitGroup{}
	stopWrite := make(chan bool, 2)
	stopRead := make(chan bool, 2)

	// write message (request => response)
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn, ch gtype.SocketChannel) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("forward response write error:", err)
			}
			stopRead <- true
		}()

		for {
			select {
			case <-stopWrite:
				return
			case <-ch.IsStop():
				return
			case data, ok := <-ch.ReadData():
				if !ok {
					return
				}

				err := conn.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					return
				}
			}
		}
	}(waitGroup, websocketConn, responseChannel)

	// read message (request <= response)
	waitGroup.Add(1)
	go func(wg *sync.WaitGroup, conn *websocket.Conn) {
		defer wg.Done()
		defer func() {
			if err := recover(); err != nil {
				s.LogError("forward response read error:", err)
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

				if msgType == websocket.BinaryMessage {
					if !s.chs.forward.WriteData(msgContent, requestId) {
						return
					}
				}
			}
		}
	}(waitGroup, websocketConn)

	waitGroup.Wait()
}

func (s *Controller) NodeForwardResponseDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "节点服务")
	function := catalog.AddFunction(method, uri, "转发响应")
	function.SetNote("接收或发送节点的转发数据，该接口保持阻塞至连接关闭")
	function.SetRemark("该接口需要客户端证书")
	function.AddInputQuery(true, "request", "请求ID", "")
	function.AddInputQuery(true, "response", "响应ID", "")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 2})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
