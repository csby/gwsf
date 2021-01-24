package gopt

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Handler interface {
	Init(router gtype.Router, api func(path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection))
}

func NewHandler(log gtype.Log, cfg *gcfg.Config, webPrefix, apiPrefix, appPrefix string) Handler {
	instance := &innerHandler{cfg: cfg}
	instance.SetLog(log)

	instance.wepPath = &gtype.Path{Prefix: webPrefix}
	instance.apiPath = &gtype.Path{Prefix: apiPrefix}
	instance.appPath = &gtype.Path{Prefix: appPrefix}

	return instance
}

type innerHandler struct {
	gtype.Base

	cfg *gcfg.Config

	wepPath *gtype.Path
	apiPath *gtype.Path
	appPath *gtype.Path
}

func (s *innerHandler) Init(router gtype.Router, api func(path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection)) {

	if api != nil {
		api(s.apiPath, nil, nil)
	}
}
