package main

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
)

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

func (s *Controller) CloudPreHandle(ctx gtype.Context, ps gtype.Params) {
	schema := ctx.Schema()
	if schema != "https" {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("protocol '%s' not support, 'https' is required", schema))
		ctx.SetHandled(true)
		return
	}

	crt := ctx.Certificate().Client
	if crt == nil {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("certificate is required"))
		ctx.SetHandled(true)
		return
	}
	ou := crt.OrganizationalUnit()
	if len(ou) < 1 {
		ctx.Error(gtype.ErrNotSupport, fmt.Errorf("organization unit of certificate is empty"))
		ctx.SetHandled(true)
		return
	}
}

func (s *Controller) CloudHello(ctx gtype.Context, ps gtype.Params) {
	ctx.Success("Hello")
}

func (s *Controller) CloudHelloDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := doc.AddCatalog("云服务").AddChild("Restful API")
	function := catalog.AddFunction(method, uri, "Hello")
	function.SetNote("示例接口，总是返回 'Hello'")
	function.SetRemark("该接口需要客户端证书")
	function.SetOutputDataExample("Hello")
}
