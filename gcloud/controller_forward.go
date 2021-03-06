package gcloud

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"time"
)

func (s *Controller) NodeForwardRequest(ctx gtype.Context, ps gtype.Params) {
	sourceInstanceId := ctx.Query("instance")
	if len(sourceInstanceId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("instance id of source note is empty"))
		return
	}
	certificateId := ctx.Query("node")
	if len(certificateId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("certificate id of target note is empty"))
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

	targetNode := s.chs.node.OnlineNode(gtype.NodeId{Certificate: certificateId})
	if targetNode == nil {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("target node(id=%s) not exist", certificateId))
		return
	}

	sourceNode := s.chs.node.OnlineNode(gtype.NodeId{Instance: sourceInstanceId})
	if sourceNode == nil {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("source node(id=%s) not exist", sourceInstanceId))
		return
	}

	conn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	channel := &ForwardChannel{
		ID:         gtype.NewGuid(),
		SrcConn:    conn,
		DstConn:    make(chan *websocket.Conn, 1),
		Error:      make(chan error, 2),
		StartTime:  gtype.DateTime(time.Now()),
		SourceNode: sourceNode,
		TargetNode: targetNode,
		TargetAddr: targetAddr,
		TargetPort: targetPort,
	}
	s.fwdChannels.Add(channel)
	defer s.fwdChannels.Del(channel.ID)
	defer s.goCloseConn(conn)

	s.writeNodeSocketMessage(targetNode.ID.Instance, gtype.WSNodeForwardTcpStart, &gtype.ForwardRequest{
		ForwardId:      gtype.ForwardId{ID: channel.ID},
		NodeInstanceID: targetNode.ID.Instance,
		TargetAddress:  targetAddr,
		TargetPort:     targetPort,
	})

	fwdInfo := &gtype.ForwardInfo{}
	channel.CopyTo(fwdInfo)
	s.writeOptSocketMessage(gtype.WSNodeForwardTcpStart, fwdInfo)
	defer s.writeOptSocketMessage(gtype.WSNodeForwardTcpEnd, &fwdInfo.ForwardId)

	select {
	case dstConn := <-channel.DstConn:
		defer s.goCloseConn(dstConn)
		ch := make(chan error, 1)
		go s.connCopy(ch, dstConn, channel.SrcConn)
		go s.connCopy(ch, channel.SrcConn, dstConn)
		err = <-ch
	case <-time.After(time.Minute):
		err = fmt.Errorf("fwd cloud timeout")
	}

	channel.Error <- err
}

func (s *Controller) NodeForwardRequestDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "结点服务")
	function := catalog.AddFunction(method, uri, "转发请求")
	function.SetNote("接收或发送结点的转发数据，该接口保持阻塞至连接关闭")
	function.SetRemark("该接口需要客户端证书")
	function.AddInputQuery(true, "instance", "发起结点实例ID", "")
	function.AddInputQuery(true, "node", "目标结点证书ID", "")
	function.AddInputQuery(true, "addr", "目标地址", "")
	function.AddInputQuery(true, "port", "目标端口", "")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 2})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Controller) NodeForwardResponse(ctx gtype.Context, ps gtype.Params) {
	forwardId := ctx.Query("id")
	if len(forwardId) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("id of forward is empty"))
		return
	}

	conn, err := s.wsGrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	channel := s.fwdChannels.Get(forwardId)
	if channel == nil {
		conn.Close()
		return
	}

	channel.DstConn <- conn

	err = <-channel.Error
}

func (s *Controller) NodeForwardResponseDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "结点服务")
	function := catalog.AddFunction(method, uri, "转发响应")
	function.SetNote("接收或发送结点的转发数据，该接口保持阻塞至连接关闭")
	function.SetRemark("该接口需要客户端证书")
	function.AddInputQuery(true, "id", "转发ID", "")
	function.SetInputExample(&gtype.SocketMessage{ID: 1})
	function.SetOutputExample(&gtype.SocketMessage{ID: 2})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Controller) readNodeMessage(message *gtype.SocketMessage, channel gtype.SocketChannel) {
	if message == nil {
		return
	}

	if message.Data == nil {
		return
	}

	if message.ID == gtype.WSNodeForwardUdpRequest {
		fwd := &gtype.ForwardUdpRequest{}
		err := message.GetData(fwd)
		if err == nil {
			go s.forwardUdpRequest(fwd)
		}
	} else if message.ID == gtype.WSNodeForwardUdpResponse {
		fwd := &gtype.ForwardUdpResponse{}
		err := message.GetData(fwd)
		if err == nil {
			go s.forwardUdpResponse(fwd)
		}
	}
}

func (s *Controller) forwardUdpRequest(request *gtype.ForwardUdpRequest) {
	if nil == request {
		return
	}

	if len(request.TargetNodeCertificateID) < 1 {
		return
	}
	targetNode := s.chs.node.OnlineNode(gtype.NodeId{Certificate: request.TargetNodeCertificateID})
	if targetNode == nil {
		return
	}
	request.TargetNodeInstanceID = targetNode.ID.Instance

	s.writeNodeSocketMessage(targetNode.ID.Instance, gtype.WSNodeForwardUdpRequest, request)
}

func (s *Controller) forwardUdpResponse(response *gtype.ForwardUdpResponse) {
	if nil == response {
		return
	}

	if len(response.SourceNodeInstanceID) < 1 {
		return
	}

	s.writeNodeSocketMessage(response.SourceNodeInstanceID, gtype.WSNodeForwardUdpResponse, response)
}
