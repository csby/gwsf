package gnode

import (
	"cdm/cdmp/data/model"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

func (s *innerController) GetNodeFwdInputSvcState(ctx gtype.Context, ps gtype.Params) {
	if s.node == nil {
		ctx.Error(gtype.ErrInternal)
		return
	}

	fwd := s.node.Forward()
	if fwd == nil {
		ctx.Error(gtype.ErrInternal)
		return
	}

	ctx.Success(model.NodeFwdInputState{
		IsRunning: fwd.IsRunning(),
	})
}

func (s *innerController) GetNodeFwdInputSvcStateDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "获取运行状态")
	function.SetNote("获取转发运行状态: true-运行中; false-已停止")
	function.SetOutputDataExample(&model.NodeFwdInputState{IsRunning: true})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *innerController) GetNodeFwdEnable(ctx gtype.Context, ps gtype.Params) {
	ctx.Success(s.cfg.Node.Forward.Enable)
}

func (s *innerController) GetNodeFwdEnableDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "获取启用状态")
	function.SetNote("获取转发启用状态: true-启用; false-禁用")
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *innerController) SetNodeFwdEnable(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.FwdEnable{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if argument.Enable == s.cfg.Node.Forward.Enable {
		ctx.Success(true)
		return
	}

	if s.cfg.Load == nil {
		ctx.Error(gtype.ErrInternal, "load not config")
		return
	}
	if s.cfg.Save == nil {
		ctx.Error(gtype.ErrInternal, "save not config")
		return
	}

	cfg, ce := s.cfg.Load()
	if ce != nil {
		ctx.Error(gtype.ErrInternal, "load config fail: ", ce)
		return
	}
	cfg.Node.Forward.Enable = argument.Enable
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "save config fail: ", err)
		return
	}
	s.cfg.Node.Forward.Enable = argument.Enable
	ctx.Success(argument.Enable)

	if argument.Enable == false {
		go s.stopNodeForward()
	}
}

func (s *innerController) SetNodeFwdEnableDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "设置启用状态")
	function.SetNote("设置转发启用状态，成功时返回true")
	function.SetRemark("禁用时将关闭当前所有转发服务，启用时仅将状态设置为启用，不会启动转发服务")
	function.SetInputJsonExample(&gcfg.FwdEnable{
		Enable: true,
	})
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *innerController) StartNodeFwd(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.Node.Enabled == false {
		ctx.Error(gtype.ErrNotSupport, "节点状态为禁用")
		return
	}
	if s.cfg.Node.Forward.Enable == false {
		ctx.Error(gtype.ErrNotSupport, "转发状态为禁用")
		return
	}

	s.startNodeForward()
	ctx.Success(true)
}

func (s *innerController) StartNodeFwdDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "启动服务")
	function.SetNote("启动转发监听服务")
	function.SetRemark("转发状态必须为启用，且存在有效的转发项目")
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrNotSupport)
}

func (s *innerController) StopNodeFwd(ctx gtype.Context, ps gtype.Params) {
	if s.cfg.Node.Enabled == false {
		ctx.Error(gtype.ErrNotSupport, "节点状态为禁用")
		return
	}
	if s.cfg.Node.Forward.Enable == false {
		ctx.Error(gtype.ErrNotSupport, "转发状态为禁用")
		return
	}

	s.stopNodeForward()
	ctx.Success(true)
}

func (s *innerController) StopNodeFwdDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "停止服务")
	function.SetNote("停止转发监听服务")
	function.SetRemark("转发状态必须为启用")
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrNotSupport)
}

func (s *innerController) onNodeFwdInputListenStateChanged(id string, isRunning bool, lastError string, newCount, oldCount int) {
	s.writeOptMessage(gtype.WSNodeFwdInputListenItemState, gtype.ForwardItemState{
		ForwardId: gtype.ForwardId{
			ID: id,
		},
		ForwardState: gtype.ForwardState{
			IsRunning: isRunning,
			LastError: lastError,
		},
	})

	if newCount == 0 {
		s.writeOptMessage(gtype.WSNodeFwdInputListenSvcState, gtype.ForwardState{
			IsRunning: false,
		})
	} else if newCount == 1 && newCount > oldCount {
		s.writeOptMessage(gtype.WSNodeFwdInputListenSvcState, gtype.ForwardState{
			IsRunning: true,
		})
	}
}
