package gtype

type Handler interface {
	InitRouting(router Router)
	BeforeRouting(ctx Context)
	AfterInput(ctx Context)
	AfterRouting(ctx Context)
	Serve(ctx Context)
	ExtendOptSetup(opt Option)
	ExtendOptApi(router Router, path *Path, preHandle HttpHandle, opt Opt)
}
