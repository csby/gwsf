package gcloud

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Handler interface {
	Init(router gtype.Router, apiExtend func(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, chs *Channels))

	OnlineNodes() []*gtype.Node
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, optChannels gtype.SocketChannelCollection) Handler {
	instance := &innerHandler{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.chs = &Channels{
		opt:          optChannels,
		onlineNodes:  gtype.NewSocketChannelCollection(),
		fwdRequests:  gtype.NewSocketChannelCollection(),
		fwdResponses: gtype.NewSocketChannelCollection(),
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

func (s *innerHandler) Init(router gtype.Router, apiExtend func(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, chs *Channels)) {
	s.controller = NewController(s.GetLog(), s.cfg, s.chs)

	path := &gtype.Path{
		Prefix: ApiPath,
	}
	preHandle := s.controller.preHandle

	router.POST(path.Uri("/node/list/online"), preHandle,
		s.controller.GetOnlineNodes, s.controller.GetOnlineNodesDoc)

	router.GET(path.Uri("/node/connect").SetIsWebsocket(true), preHandle,
		s.controller.NodeConnect, s.controller.NodeConnectDoc)

	if apiExtend != nil {
		apiExtend(router, path, preHandle, s.chs)
	}
}

func (s *innerHandler) OnlineNodes() []*gtype.Node {
	return s.chs.onlineNodes.OnlineNodes()
}
