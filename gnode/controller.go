package gnode

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type innerController struct {
	gtype.Base

	node Handler
	cfg  *gcfg.Config
	wcs  gtype.SocketChannelCollection

	nodeInstance string // 节点实例ID
	nodeId       string // 节点证书标识ID
	nodeName     string // 节点名称
}

func (s *innerController) initRouter(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle) {
	if router == nil || path == nil {
		return
	}

	// 获取启用状态
	router.POST(path.Uri("/node/info"), preHandle,
		s.GetNodeInfo, s.GetNodeInfoDoc)
	router.POST(path.Uri("/cfg/node/enable/get"), preHandle,
		s.GetNodeEnable, s.GetNodeEnableDoc)
	// 获取在线状态
	router.POST(path.Uri("/node/online/state"), preHandle,
		s.GetNodeOnlineState, s.GetNodeOnlineStateDoc)
	// 获取在线目标节点
	router.POST(path.Uri("/node/online/target/list"), preHandle,
		s.GetNodeOnlineTargets, s.GetNodeOnlineTargetsDoc)
	// 转发-获取运行状态
	router.POST(path.Uri("/node/fwd/input/listen/state"), preHandle,
		s.GetNodeFwdInputSvcState, s.GetNodeFwdInputSvcStateDoc)
	// 转发-获取启用状态
	router.POST(path.Uri("/cfg/node/fwd/enable/get"), preHandle,
		s.GetNodeFwdEnable, s.GetNodeFwdEnableDoc)
	// 转发-设置启用状态
	router.POST(path.Uri("/cfg/node/fwd/enable/set"), preHandle,
		s.SetNodeFwdEnable, s.SetNodeFwdEnableDoc)
	// 转发-启动服务
	router.POST(path.Uri("/cfg/node/fwd/input/svc/listen/start"), preHandle,
		s.StartNodeFwd, s.StartNodeFwdDoc)
	// 转发-停止服务
	router.POST(path.Uri("/cfg/node/fwd/input/svc/listen/stop"), preHandle,
		s.StopNodeFwd, s.StopNodeFwdDoc)
	// 转发-获取转发列表
	router.POST(path.Uri("/cfg/node/fwd/item/list"), preHandle,
		s.ListNoteFwd, s.ListNoteFwdDoc)
	// 转发-添加转发项目
	router.POST(path.Uri("/cfg/node/fwd/item/add"), preHandle,
		s.AddNoteFwd, s.AddNoteFwdDoc)
	// 转发-修改转发项目
	router.POST(path.Uri("/cfg/node/fwd/item/mod"), preHandle,
		s.ModNoteFwd, s.ModNoteFwdDoc)
	// 转发-删除转发项目
	router.POST(path.Uri("/cfg/node/fwd/item/del"), preHandle,
		s.DelNoteFwd, s.DelNoteFwdDoc)
}

func (s *innerController) createCatalog(doc gtype.Doc, names ...string) gtype.Catalog {
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

func (s *innerController) writeOptMessage(id int, data interface{}) bool {
	if s.wcs == nil {
		return false
	}

	msg := &gtype.SocketMessage{
		ID:   id,
		Data: data,
	}

	s.wcs.Write(msg, nil)

	return true
}
