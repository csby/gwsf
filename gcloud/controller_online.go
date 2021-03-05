package gcloud

import (
	"github.com/csby/gwsf/gtype"
	"time"
)

func (s *Controller) GetOnlineNodes(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.chs.node.OnlineNodes())
}

func (s *Controller) GetOnlineNodesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "结点服务")
	function := catalog.AddFunction(method, uri, "获取在线结点")
	function.SetNote("获取当前所有在线列表")
	function.SetRemark("该接口需要客户端证书")
	function.SetOutputDataExample([]*gtype.Node{
		{
			ID: gtype.NodeId{
				Instance:    gtype.NewGuid(),
				Certificate: gtype.NewGuid(),
			},
			Kind:      "client",
			Name:      "测试结点",
			IP:        "192.168.1.100",
			LoginTime: gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}
