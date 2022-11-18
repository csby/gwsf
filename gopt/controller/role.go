package controller

import (
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

type Role struct {
	controller

	isCluster bool
	isCloud   bool
	isNode    bool
}

func NewRole(log gtype.Log, cfg *gcfg.Config, isCluster, isCloud, isNode bool) *Role {
	instance := &Role{}
	instance.SetLog(log)
	instance.cfg = cfg
	instance.isCluster = isCluster
	instance.isCloud = isCloud
	instance.isNode = isNode

	return instance
}

func (s *Role) GetServerRole(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(&gtype.ServerRole{
		Cluster: s.isCluster,
		Cloud:   s.isCloud,
		Node:    s.isNode,
		Service: s.cfg.Sys.Svc.Enabled,
		Proxy:   s.cfg.ReverseProxy.Enabled,
	})
}

func (s *Role) GetServerRoleDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "系统角色")
	function := catalog.AddFunction(method, uri, "获取服务角色")
	function.SetNote("获取当前服务角色相关信息")
	function.SetOutputDataExample(&gtype.ServerRole{
		Cloud: false,
		Node:  true,
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
