package gcloud

import (
	"github.com/csby/gwsf/gtype"
	"time"
)

func (s *Controller) GetNodes(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.clients.Items())
}

func (s *Controller) GetNodesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, "云端服务")
	function := catalog.AddFunction(method, uri, "获取所有节点")
	function.SetNote("获取当前所有注册的节点列表")
	function.SetOutputDataExample([]*Node{
		{
			Id:      gtype.NewGuid(),
			Kind:    "client",
			Name:    "测试节点",
			RegIp:   "192.168.1.100",
			RegTime: gtype.DateTime(time.Now()),
			Instances: []*NodeInstance{{
				Id:   gtype.NewGuid(),
				Time: gtype.DateTime(time.Now()),
			}},
		},
	})
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}

func (s *Controller) GetOnlineNodesDocForOpt(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, "云端服务")
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

func (s *Controller) ModNode(ctx gtype.Context, ps gtype.Params) {
	argument := &NodeModify{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput.SetDetail(err))
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput.SetDetail("节点标识(id)为空"))
		return
	}

	err = s.clients.NodeModify(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput.SetDetail(err))
		return
	}

	ctx.Success(nil)
}

func (s *Controller) ModNodeDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, "云端服务")
	function := catalog.AddFunction(method, uri, "修改节点信息")
	function.SetNote("修改指定节点信息")
	function.SetInputJsonExample(&NodeModify{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}

func (s *Controller) DelNode(ctx gtype.Context, ps gtype.Params) {
	argument := &NodeDelete{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput.SetDetail(err))
		return
	}
	if len(argument.Id) < 1 {
		ctx.Error(gtype.ErrInput.SetDetail("节点标识(id)为空"))
		return
	}

	ge := s.clients.NodeDelete(argument.Id)
	if ge != nil {
		ctx.Error(ge)
		return
	}

	ctx.Success(nil)
}

func (s *Controller) DelNodeDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, "云端服务")
	function := catalog.AddFunction(method, uri, "删除节点")
	function.SetNote("删除已注册的节点")
	function.SetRemark("仅能删除离线的节点，在线节点则不能删除")
	function.SetInputJsonExample(&NodeDelete{
		Id: gtype.NewGuid(),
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
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
