package gnode

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

const (
	nodeCatalogRoot = "节点服务"
	nodeCatalogFwd  = "端口转发"
)

type Node struct {
	ID   string `json:"id" note:"证书标识ID"`
	Name string `json:"name" note:" 结点名称"`
}

type NodeOnlineState struct {
	IsOnline bool `json:"online" note:"状态: true-在线; false-离线"`
}

func (s *innerController) GetNodeInfo(ctx gtype.Context, ps gtype.Params) {
	info := &gtype.NodeInfo{
		NodeId: gtype.NodeId{
			Instance:    s.nodeInstance,
			Certificate: s.nodeId,
		},
		Name: s.nodeName,
	}
	if s.cfg != nil {
		info.Enabled = s.cfg.Node.Enabled
		info.CloudAddress = fmt.Sprintf("%s:%d", s.cfg.Node.CloudServer.Address, s.cfg.Node.CloudServer.Port)
	}
	if s.node != nil {
		cloud := s.node.Cloud()
		if cloud != nil {
			info.Online = cloud.IsConnected()
		}
	}

	ctx.Success(info)
}

func (s *innerController) GetNodeInfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot)
	function := catalog.AddFunction(method, uri, "获取节点信息")
	function.SetNote("获取节点当前状态信息")
	function.SetOutputDataExample(&gtype.NodeInfo{})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *innerController) GetNodeEnable(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.cfg.Node.Enabled)
}

func (s *innerController) GetNodeEnableDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot)
	function := catalog.AddFunction(method, uri, "获取启用状态")
	function.SetNote("获取节点启用状态: true-启用; false-禁用")
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *innerController) GetNodeOnlineState(ctx gtype.Context, ps gtype.Params) {
	if s.node == nil {
		ctx.Error(gtype.ErrInternal)
		return
	}

	cloud := s.node.Cloud()
	if cloud == nil {
		ctx.Error(gtype.ErrInternal)
		return
	}

	ctx.Success(NodeOnlineState{
		IsOnline: cloud.IsConnected(),
	})
}

func (s *innerController) GetNodeOnlineStateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot)
	function := catalog.AddFunction(method, uri, "获取在线状态")
	function.SetNote("获取节点在线状态: true-在线; false-离线")
	function.SetOutputDataExample(&NodeOnlineState{
		IsOnline: false,
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *innerController) GetNodeOnlineTargets(ctx gtype.Context, ps gtype.Params) {
	if s.node == nil {
		ctx.Error(gtype.ErrInternal)
		return
	}

	cloud := s.node.Cloud()
	if cloud == nil {
		ctx.Error(gtype.ErrInternal)
		return
	}

	result := cloud.PostJson("/node/list/online", nil)
	if result.Code != 0 {
		ctx.Error(gtype.ErrInternal.SetDetail(result.Error.Detail))
		return
	}
	items := make([]*gtype.Node, 0)
	err := result.GetData(&items)
	if err != nil {
		ctx.Error(gtype.ErrInternal.SetDetail(err))
		return
	}

	nodes := make([]*Node, 0)
	keys := make(map[string]byte, 0)
	count := len(items)
	for index := 0; index < count; index++ {
		item := items[index]
		if item == nil {
			continue
		}
		if _, ok := keys[item.ID.Certificate]; ok {
			continue
		}
		keys[item.ID.Certificate] = 0

		node := &Node{
			ID:   item.ID.Certificate,
			Name: item.Name,
		}
		nodes = append(nodes, node)
	}

	ctx.Success(nodes)
}

func (s *innerController) GetNodeOnlineTargetsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot)
	function := catalog.AddFunction(method, uri, "获取在线目标节点")
	function.SetNote("获取在线目标节点列表")
	function.SetOutputDataExample([]*Node{
		{
			ID:   gtype.NewGuid(),
			Name: "测试节点",
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *innerController) onNodeOnlineStateChanged(isConnected bool) {
	s.writeOptMessage(gtype.WSNodeOnlineStateChanged, NodeOnlineState{
		IsOnline: isConnected,
	})
}

func (s *innerController) stopNodeForward() {
	if s.node == nil {
		return
	}

	fwd := s.node.Forward()
	if fwd == nil {
		return
	}

	fwd.Stop()
}

func (s *innerController) startNodeForward() {
	if s.node == nil {
		return
	}

	fwd := s.node.Forward()
	if fwd == nil {
		return
	}

	fwd.Start()
}

func (s *innerController) getNodeForwardStates(states map[string]*gcfg.FwdState) {
	if s.node == nil {
		return
	}

	fwd := s.node.Forward()
	if fwd == nil {
		return
	}

	fwd.GetStates(states)
}
