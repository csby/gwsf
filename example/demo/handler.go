package main

import (
	"github.com/csby/gwsf/gcloud"
	"github.com/csby/gwsf/gnode"
	"github.com/csby/gwsf/gtype"
	"net/http"
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
	s.apiController.cloudHandler = s.cloudHandler
	router.POST(apiPath.Uri("/hello"), nil,
		s.apiController.Hello, s.apiController.HelloDoc)
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	method := ctx.Method()
	// enable across access
	if method == http.MethodOptions {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
		return
	}
}

func (s *Handler) AfterInput(ctx gtype.Context) {

}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) ExtendOptSetup(opt gtype.Option) {
	if opt == nil {
		return
	}

	opt.SetCloud(cfg.Cloud.Enabled)
	opt.SetNode(cfg.Node.Enabled)
}

func (s *Handler) ExtendOptApi(router gtype.Router,
	path *gtype.Path,
	preHandle gtype.HttpHandle,
	opt gtype.Opt) {
	var wsc gtype.SocketChannelCollection = nil
	if opt != nil {
		wsc = opt.Wsc()
	}
	s.optSocketChannels = wsc

	s.cloudHandler = gcloud.NewHandler(s.GetLog(), &cfg.Config, wsc)
	s.nodeHandler = gnode.NewHandler(s.GetLog(), &cfg.Config, wsc)

	s.cloudHandler.Init(router, path, preHandle, nil)
	s.nodeHandler.Init(router, path, preHandle)

	router.POST(path.Uri("/node/list/online"), preHandle,
		s.apiController.GetOnlineNodes, s.apiController.GetOnlineNodesDoc)
	router.POST(path.Uri("/forward/list/online"), preHandle,
		s.apiController.GetForwardNodes, s.apiController.GetForwardNodesDoc)
}
