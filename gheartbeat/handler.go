package gheartbeat

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Handler interface {
	Init(router gtype.Router, optPath *gtype.Path, optPreHandle gtype.HttpHandle)
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, opt gtype.SocketChannelCollection) Handler {
	instance := &innerHandler{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.controller = NewController(log, cfg, opt)

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg        *gcfg.Config
	controller *Controller
}

func (s *innerHandler) Init(router gtype.Router, optPath *gtype.Path, optPreHandle gtype.HttpHandle) {

	path := &gtype.Path{
		Prefix: ApiPath,
	}

	router.GET(path.Uri("/connect").SetIsWebsocket(true), nil,
		s.controller.Connect, s.controller.ConnectDoc)

	if optPath != nil {
		s.initOpt(router, optPath, optPreHandle)
	}

}

func (s *innerHandler) initOpt(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle) {
	if router == nil || path == nil {
		return
	}

	router.POST(path.Uri("/heartbeat/conn/list"), preHandle,
		s.controller.GetList, s.controller.GetListDoc)
}
