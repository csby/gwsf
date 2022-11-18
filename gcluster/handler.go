package gcluster

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Handler interface {
	Init(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle,
		apiExtend func(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, chs *Channels))
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, opt gtype.SocketChannelCollection) Handler {
	instance := &innerHandler{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.chs = &Channels{
		opt:     opt,
		cluster: gtype.NewSocketChannelCollection(),
	}
	instance.controller = NewController(log, cfg, instance.chs)

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg        *gcfg.Config
	chs        *Channels
	controller *Controller
}

func (s *innerHandler) Init(router gtype.Router, optPath *gtype.Path, optPreHandle gtype.HttpHandle,
	apiExtend func(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, chs *Channels)) {

	path := &gtype.Path{
		Prefix: ApiPath,
	}
	preHandle := s.controller.preHandle

	router.GET(path.Uri("/instance/connect").SetIsWebsocket(true), preHandle,
		s.controller.NodeConnect, s.controller.NodeConnectDoc)

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

	router.POST(path.Uri("/cluster/info"), preHandle,
		s.controller.GetClusterInfo, s.controller.GetClusterInfoDoc)
	router.POST(path.Uri("/cluster/msg/sync/send"), preHandle,
		s.controller.SendSyncMessage, s.controller.SendSyncMessageDoc)
}
