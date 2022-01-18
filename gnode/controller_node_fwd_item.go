package gnode

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

func (s *innerController) ListNoteFwd(ctx gtype.Context, ps gtype.Params) {
	results := make([]*gcfg.FwdInfo, 0)

	states := make(map[string]*gcfg.FwdState, 0)
	items := s.cfg.Node.Forward.Items
	count := len(items)
	for index := 0; index < count; index++ {
		item := items[index]
		if item == nil {
			continue
		}
		result := &gcfg.FwdInfo{}
		result.CopyFrom(item)
		results = append(results, result)

		if len(result.ID) > 0 {
			states[result.ID] = &result.FwdState
		}
	}
	s.getNodeForwardStates(states)

	ctx.Success(results)
}

func (s *innerController) ListNoteFwdDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "获取转发列表")
	function.SetNote("获取获取转发列表信息")
	function.SetOutputDataExample([]*gcfg.FwdInfo{
		{
			Fwd: gcfg.Fwd{
				FwdId: gcfg.FwdId{
					ID: gtype.NewGuid(),
				},
				FwdContent: gcfg.FwdContent{
					FwdEnable: gcfg.FwdEnable{
						Enable: true,
					},
					Protocol:       "tcp",
					ListenAddress:  "127.0.0.1",
					ListenPort:     8081,
					TargetNodeID:   gtype.NewGuid(),
					TargetNodeName: "目标节点名称",
					TargetAddress:  "192.168.1.8",
					TargetPort:     80,
				},
			},
			FwdState: gcfg.FwdState{
				IsRunning: true,
			},
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *innerController) AddNoteFwd(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.Fwd{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if _, ok := s.cfg.Node.Forward.GetItemId(argument.ListenAddress, argument.ListenPort); ok {
		ctx.Error(gtype.ErrInput, fmt.Errorf("监听地址'%s:%d'已存在", argument.ListenAddress, argument.ListenPort))
		return
	}
	argument.ID = gtype.NewGuid()

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
	if cfg.Node.Forward.Items == nil {
		cfg.Node.Forward.Items = make([]*gcfg.Fwd, 0)
	}
	cfg.Node.Forward.Items = append(cfg.Node.Forward.Items, argument)
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "save config fail: ", err)
		return
	}

	if s.cfg.Node.Forward.Items == nil {
		s.cfg.Node.Forward.Items = make([]*gcfg.Fwd, 0)
	}
	s.cfg.Node.Forward.Items = append(s.cfg.Node.Forward.Items, argument)

	ctx.Success(argument.ID)
}

func (s *innerController) AddNoteFwdDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "添加转发项目")
	function.SetNote("添加转发项目，成功时返回项目标识ID")
	function.SetInputJsonExample(&gcfg.FwdContent{
		FwdEnable: gcfg.FwdEnable{
			Enable: true,
		},
		Protocol:       "tcp",
		ListenAddress:  "127.0.0.1",
		ListenPort:     8081,
		TargetNodeID:   gtype.NewGuid(),
		TargetNodeName: "目标节点名称",
		TargetAddress:  "192.168.1.8",
		TargetPort:     80,
	})
	function.SetOutputDataExample(gtype.NewGuid())
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *innerController) ModNoteFwd(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.Fwd{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
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
	if cfg.Node.Forward.Items == nil {
		cfg.Node.Forward.Items = make([]*gcfg.Fwd, 0)
	}
	item := cfg.Node.Forward.GetItem(argument.ID)
	if item == nil {
		ctx.Error(gtype.ErrInput, fmt.Errorf("id=%s 不存在", argument.ID))
		return
	}
	if id, ok := cfg.Node.Forward.GetItemId(argument.ListenAddress, argument.ListenPort); ok && id != argument.ID {
		ctx.Error(gtype.ErrInput, fmt.Errorf("监听地址'%s:%d'已存在", argument.ListenAddress, argument.ListenPort))
		return
	}

	argument.CopyTo(item)
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "save config fail: ", err)
		return
	}

	item = s.cfg.Node.Forward.GetItem(argument.ID)
	if item != nil {
		argument.CopyTo(item)
	}

	ctx.Success(true)
}

func (s *innerController) ModNoteFwdDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "修改转发项目")
	function.SetNote("修改转发项目，成功时返回true")
	function.SetInputJsonExample(&gcfg.Fwd{
		FwdId: gcfg.FwdId{
			ID: gtype.NewGuid(),
		},
		FwdContent: gcfg.FwdContent{
			FwdEnable: gcfg.FwdEnable{
				Enable: true,
			},
			Protocol:       "tcp",
			ListenAddress:  "127.0.0.1",
			ListenPort:     8081,
			TargetNodeID:   gtype.NewGuid(),
			TargetNodeName: "目标节点名称",
			TargetAddress:  "192.168.1.8",
			TargetPort:     80,
		},
	})
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
}

func (s *innerController) DelNoteFwd(ctx gtype.Context, ps gtype.Params) {
	argument := &gcfg.Fwd{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}

	cfg, ce := s.cfg.Load()
	if ce != nil {
		ctx.Error(gtype.ErrInternal, "load config fail: ", ce)
		return
	}
	items := cfg.Node.Forward.Items
	count := len(items)
	newItems := make([]*gcfg.Fwd, 0)
	for index := 0; index < count; index++ {
		item := items[index]
		if item == nil {
			continue
		}
		if argument.ID == item.ID {
			continue
		}
		newItems = append(newItems, item)
	}
	if len(newItems) >= count {
		ctx.Error(gtype.ErrInput, fmt.Errorf("id=%s 不存在", argument.ID))
		return
	}
	cfg.Node.Forward.Items = newItems
	err = s.cfg.Save(cfg)
	if err != nil {
		ctx.Error(gtype.ErrInternal, "save config fail: ", err)
		return
	}

	s.cfg.Node.Forward.Items = newItems
	ctx.Success(true)
}

func (s *innerController) DelNoteFwdDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, nodeCatalogRoot, nodeCatalogFwd)
	function := catalog.AddFunction(method, uri, "删除转发项目")
	function.SetNote("删除转发项目，成功时返回true")
	function.SetInputJsonExample(&gcfg.FwdId{
		ID: gtype.NewGuid(),
	})
	function.SetOutputDataExample(true)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
	function.AddOutputError(gtype.ErrInternal)
}
