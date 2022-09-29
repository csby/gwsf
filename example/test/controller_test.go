package test

import "github.com/csby/gwsf/gtype"

type Controller struct {
	gtype.Base
}

func (s *Controller) Hello(ctx gtype.Context, ps gtype.Params) {
	ctx.Success("Hello")

	crt := ctx.Certificate().Client
	if crt != nil {
		s.LogInfo("client certificate: sn=", crt.SerialNumberString(), ", ou=", crt.OrganizationalUnit())
	}
}

func (s *Controller) HelloDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := doc.AddCatalog("UnitTest").AddChild("Restful")
	function := catalog.AddFunction(method, uri, "Hello World")
	function.SetNote("restful api, return data with 'Hello'")
	function.SetOutputDataExample("Hello")
	function.AddOutputError(gtype.ErrInput)
}
