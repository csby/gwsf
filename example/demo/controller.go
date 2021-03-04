package main

import (
	"github.com/csby/gwsf/gcloud"
	"github.com/csby/gwsf/gtype"
	"time"
)

type Controller struct {
	gtype.Base

	cloudHandler gcloud.Handler
}

func (s *Controller) Hello(ctx gtype.Context, ps gtype.Params) {
	ctx.Success("Hello")
}

func (s *Controller) HelloDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := doc.AddCatalog("服务示例").AddChild("Restful API")
	function := catalog.AddFunction(method, uri, "Hello")
	function.SetNote("示例接口，总是返回 'Hello'")
	function.SetRemark("该接口不需要凭证")
	function.SetOutputDataExample("Hello")
}

func (s *Controller) GetOnlineNodes(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.cloudHandler.OnlineNodes())
}

func (s *Controller) GetOnlineNodesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := doc.AddCatalog("管理平台接口").AddChild("节点管理")
	function := catalog.AddFunction(method, uri, "获取在线节点")
	function.SetNote("获取当前所有在线列表")
	function.SetOutputDataExample([]*gtype.Node{
		{
			ID: gtype.NodeId{
				Instance:    gtype.NewGuid(),
				Certificate: gtype.NewGuid(),
			},
			Kind:      "client",
			Name:      "测试节点",
			IP:        "192.168.1.100",
			LoginTime: gtype.DateTime(time.Now()),
		},
	})
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}
