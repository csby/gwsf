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
	catalog := doc.AddCatalog("管理平台接口").AddChild("云端服务").AddChild("结点管理")
	function := catalog.AddFunction(method, uri, "获取在线结点")
	function.SetNote("获取当前所有在线列表")
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
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}

func (s *Controller) GetForwardNodes(ctx gtype.Context, ps gtype.Params) {
	argument := &gtype.ForwardInfoFilter{}
	ctx.GetJson(argument)

	ctx.Success(s.cloudHandler.OnlineForwards(argument))
}

func (s *Controller) GetForwardNodesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := doc.AddCatalog("管理平台接口").AddChild("云端服务").AddChild("结点管理")
	function := catalog.AddFunction(method, uri, "获取结点转发连接")
	function.SetNote("获取当前所有正在转发的连接列表")
	function.SetInputJsonExample(&gtype.ForwardInfoFilter{})
	function.SetOutputDataExample([]*gtype.ForwardInfo{
		{
			ForwardId: gtype.ForwardId{
				ID: gtype.NewGuid(),
			},
			Time: gtype.DateTime(time.Now()),
			SourceNode: &gtype.Node{
				ID: gtype.NodeId{
					Instance:    gtype.NewGuid(),
					Certificate: gtype.NewGuid(),
				},
				Kind:      "user",
				Name:      "发起结点",
				IP:        "172.16.1.100",
				LoginTime: gtype.DateTime(time.Now()),
			},
			TargetNode: &gtype.Node{
				ID: gtype.NodeId{
					Instance:    gtype.NewGuid(),
					Certificate: gtype.NewGuid(),
				},
				Kind:      "client",
				Name:      "目标结点",
				IP:        "192.168.1.100",
				LoginTime: gtype.DateTime(time.Now()),
			},
			TargetHost: "192.168.1.7:8080",
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}
