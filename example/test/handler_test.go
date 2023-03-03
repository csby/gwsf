package test

import "github.com/csby/gwsf/gtype"

type Handler struct {
	gtype.Base

	controller *Controller
}

func (s *Handler) InitRouting(router gtype.Router) {
	s.controller = &Controller{}
	s.controller.SetLog(s.GetLog())

	router.POST(path.Uri(uriHello), nil, s.controller.Hello, s.controller.HelloDoc)
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	// enable across access
	if ctx.Method() == "OPTIONS" {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
	}
}

func (s *Handler) AfterInput(ctx gtype.Context) {

}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) Serve(ctx gtype.Context) {

}

func (s *Handler) ExtendOptSetup(opt gtype.Option) {
}

func (s *Handler) ExtendOptApi(router gtype.Router,
	path *gtype.Path,
	preHandle gtype.HttpHandle,
	opt gtype.Opt) {

}
