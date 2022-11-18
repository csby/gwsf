package gcluster

import "github.com/csby/gwsf/gtype"

func (s *Controller) GetClusterInfo(ctx gtype.Context, ps gtype.Params) {
	info := &Cluster{
		Index:  s.cfg.Cluster.Index,
		Enable: s.cfg.Cluster.Enable,
		Nodes:  make([]*Node, 0),
	}

	items := s.cfg.Cluster.Instances
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		node := &Node{
			Index:   item.Index,
			Address: item.Address,
			Port:    item.Port,
		}

		inst := s.getInstance(item.Index)
		if inst != nil {
			in := inst.In
			if in != nil {
				node.Status.In = in.Connected()
			}

			out := inst.Out
			if out != nil {
				node.Status.Out = out.Connected()
			}
		}

		info.Nodes = append(info.Nodes, node)
	}

	ctx.Success(info)
}

func (s *Controller) GetClusterInfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, "集群服务")
	function := catalog.AddFunction(method, uri, "获取集群信息")
	function.SetOutputDataExample(&Cluster{
		Nodes: []*Node{
			{},
		},
	})
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}

func (s *Controller) SendSyncMessage(ctx gtype.Context, ps gtype.Params) {
	argument := &gtype.SocketMessage{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput.SetDetail(err))
		return
	}

	go s.Write(argument, nil)

	ctx.Success(nil)
}

func (s *Controller) SendSyncMessageDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, "集群服务")
	function := catalog.AddFunction(method, uri, "发送同步信息")
	function.SetNote("将需要同步的消息发送到其他节点")
	function.SetInputJsonExample(&gtype.SocketMessage{
		ID:   0,
		Data: "测试消息",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}
