package main

import "github.com/csby/gwsf/gtype"

func NewHandler(log gtype.Log) gtype.Handler {
	instance := &Handler{}
	instance.SetLog(log)

	instance.apiController = &Controller{}
	instance.apiController.SetLog(log)

	return instance
}

type Handler struct {
	gtype.Base

	apiController *Controller
}

func (s *Handler) InitRouting(router gtype.Router) {
	router.POST(apiPath.Uri("/hello"), nil, s.apiController.Hello, s.apiController.HelloDoc)
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	// enable across access
	if ctx.Method() == "OPTIONS" {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
		return
	}
	origin := ctx.Request().Header.Get("Origin")
	s.LogDebug("Origin: ", origin)
}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) ExtendOptApi(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection) {

}
