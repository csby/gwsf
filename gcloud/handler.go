package gcloud

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Handler interface {
	Init(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle,
		apiExtend func(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, chs *Channels))

	OnlineNodes() []*gtype.Node
	OnlineForwards(filter *gtype.ForwardInfoFilter) gtype.ForwardInfoArray
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, optChannels gtype.SocketChannelCollection) Handler {
	instance := &innerHandler{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.chs = &Channels{
		opt:  optChannels,
		node: gtype.NewSocketChannelCollection(),
	}
	instance.controller = NewController(log, cfg, instance.chs)

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg *gcfg.Config
	chs *Channels

	controller *Controller
}

func (s *innerHandler) Init(router gtype.Router, optPath *gtype.Path, optPreHandle gtype.HttpHandle,
	apiExtend func(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, chs *Channels)) {

	path := &gtype.Path{
		Prefix: ApiPath,
	}
	preHandle := s.controller.preHandle

	router.POST(path.Uri("/node/service/info"), preHandle,
		s.controller.GetNodeServiceInfo, s.controller.GetNodeServiceInfoDoc)
	router.POST(path.Uri("/node/list/online"), preHandle,
		s.controller.GetOnlineNodes, s.controller.GetOnlineNodesDoc)
	router.GET(path.Uri("/node/connect").SetIsWebsocket(true), preHandle,
		s.controller.NodeConnect, s.controller.NodeConnectDoc)

	router.GET(path.Uri("/fwd/request").SetIsWebsocket(true), preHandle,
		s.controller.NodeForwardRequest, s.controller.NodeForwardRequestDoc)
	router.GET(path.Uri("/fwd/response").SetIsWebsocket(true), preHandle,
		s.controller.NodeForwardResponse, s.controller.NodeForwardResponseDoc)

	if apiExtend != nil {
		apiExtend(router, path, preHandle, s.chs)
	}

	if optPath != nil {
		s.initOpt(router, optPath, optPreHandle)
	}
}

func (s *innerHandler) initOpt(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle) {
	if router == nil || path == nil {
		return
	}

	router.POST(path.Uri("/cloud/node/list/all"), preHandle,
		s.controller.GetNodes, s.controller.GetNodesDoc)
	router.POST(path.Uri("/cloud/node/list/online"), preHandle,
		s.controller.GetOnlineNodes, s.controller.GetOnlineNodesDocForOpt)
	router.POST(path.Uri("/cloud/node/info/mod"), preHandle,
		s.controller.ModNode, s.controller.ModNodeDoc)
	router.POST(path.Uri("/cloud/node/info/del"), preHandle,
		s.controller.DelNode, s.controller.DelNodeDoc)
}

func (s *innerHandler) OnlineNodes() []*gtype.Node {
	return s.chs.node.OnlineNodes()
}

func (s *innerHandler) OnlineForwards(filter *gtype.ForwardInfoFilter) gtype.ForwardInfoArray {
	return s.controller.fwdChannels.Lst(filter)
}
