package gheartbeat

import "github.com/csby/gwsf/gtype"

func (s *Controller) GetList(ctx gtype.Context, ps gtype.Params) {
	items := s.heartbeats.Items
	ctx.Success(items)
}

func (s *Controller) GetListDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createOptCatalog(doc, ApiCatalog)
	function := catalog.AddFunction(method, uri, "获取连接列表")
	function.SetOutputDataExample([]*gtype.Heartbeat{
		{},
	})
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}
