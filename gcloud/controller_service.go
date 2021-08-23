package gcloud

import (
	"github.com/csby/gwsf/gtype"
	"time"
)

func (s *Controller) GetNodeServiceInfo(ctx gtype.Context, ps gtype.Params) {
	data := &gtype.SvcCloudInfo{BootTime: gtype.DateTime(s.bootTime)}
	cfg := s.cfg
	if cfg != nil {
		data.Name = cfg.Module.Name
		data.Version = cfg.Module.Version
		data.Remark = cfg.Module.Remark
	}
	data.ClientOU = ctx.ClientOrganization()
	crt := ctx.Certificate().Server
	if crt != nil {
		data.ServerOU = crt.OrganizationalUnit()
	}

	ctx.Success(data)
}

func (s *Controller) GetNodeServiceInfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "结点服务")
	function := catalog.AddFunction(method, uri, "获取服务信息")
	function.SetNote("获取当前节点云服务信息")
	function.SetRemark("该接口需要客户端证书")
	function.SetOutputDataExample([]*gtype.SvcCloudInfo{
		{
			Name:     "server",
			BootTime: gtype.DateTime(time.Now()),
			Version:  "1.0.1.0",
			Remark:   "XXX服务",
		},
	})
	function.AddOutputError(gtype.ErrNotSupport)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrException)
}
