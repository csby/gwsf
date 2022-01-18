package main

import (
	"github.com/csby/gwsf/gcloud"
	"github.com/csby/gwsf/gnode"
	"github.com/csby/gwsf/gtype"
)

func NewHandler(log gtype.Log) gtype.Handler {
	instance := &Handler{}
	instance.SetLog(log)

	instance.apiController = &Controller{}
	instance.apiController.SetLog(log)

	return instance
}

type Handler struct {
	gtype.Base

	optSocketChannels gtype.SocketChannelCollection
	apiController     *Controller
	cloudHandler      gcloud.Handler
	nodeHandler       gnode.Handler
}

func (s *Handler) InitRouting(router gtype.Router) {
	s.cloudHandler = gcloud.NewHandler(s.GetLog(), &cfg.Config, s.optSocketChannels)
	s.nodeHandler = gnode.NewHandler(s.GetLog(), &cfg.Config, s.optSocketChannels)

	s.cloudHandler.Init(router, nil, nil, nil)
	s.nodeHandler.Init(nil, nil, nil)

	s.apiController.cloudHandler = s.cloudHandler
	router.POST(apiPath.Uri("/hello"), nil,
		s.apiController.Hello, s.apiController.HelloDoc)
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	// enable across access
	if ctx.Method() == "OPTIONS" {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
		return
	}
}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) ExtendOptSetup(opt gtype.Option) {
}

func (s *Handler) ExtendOptApi(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection) {
	s.optSocketChannels = wsc

	router.POST(path.Uri("/node/list/online"), preHandle,
		s.apiController.GetOnlineNodes, s.apiController.GetOnlineNodesDoc)
	router.POST(path.Uri("/forward/list/online"), preHandle,
		s.apiController.GetForwardNodes, s.apiController.GetForwardNodesDoc)
}
