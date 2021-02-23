package main

import "github.com/csby/gwsf/gtype"

type Controller struct {
	gtype.Base
}

func (s *Controller) Hello(ctx gtype.Context, ps gtype.Params) {
	ctx.Success("Hello")
}

func (s *Controller) HelloDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := doc.AddCatalog("服务示例").AddChild("Restful API")
	function := catalog.AddFunction(method, uri, "Hello")
	function.SetNote("示例接口，总是返回 'Hello'")
	function.SetRemark("该接口不需要凭证")
	function.SetOutputDataExample("Hello")
}
