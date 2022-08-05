package controller

import (
	"fmt"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
)

func (s *Service) StartOther(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetOtherByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	err = s.start(argument.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: argument.Name}
	svcStatus.Status, err = s.getStatus(argument.Name)
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) StartOtherDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogOther)
	function := catalog.AddFunction(method, uri, "启动服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) StopOther(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetOtherByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	err = s.stop(argument.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: argument.Name}
	svcStatus.Status, err = s.getStatus(argument.Name)
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) StopOtherDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogOther)
	function := catalog.AddFunction(method, uri, "停止服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) RestartOther(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ServerArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInternal, "服务名称(name)为空")
		return
	}
	info := s.cfg.Sys.Svc.GetOtherByServiceName(argument.Name)
	if info == nil {
		ctx.Error(gtype.ErrInternal, fmt.Sprintf("服务(%s)不存在", argument.Name))
		return
	}

	err = s.restart(argument.Name)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	svcStatus := &gmodel.ServiceStatus{Name: argument.Name}
	svcStatus.Status, err = s.getStatus(argument.Name)
	if err == nil {
		go s.writeOptMessage(gtype.WSSvcStatusChanged, svcStatus)
	}

	ctx.Success(info)
}

func (s *Service) RestartOtherDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogOther)
	function := catalog.AddFunction(method, uri, "重启服务")
	function.SetInputJsonExample(&gmodel.ServerArgument{
		Name: "example",
	})
	function.SetOutputDataExample(nil)
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Service) GetOthers(ctx gtype.Context, ps gtype.Params) {
	results := make([]*gmodel.ServiceOtherInfo, 0)

	if s.cfg != nil {
		items := s.cfg.Sys.Svc.Others
		c := len(items)
		for i := 0; i < c; i++ {
			item := items[i]
			if item == nil {
				continue
			}
			if len(item.ServiceName) < 1 {
				continue
			}

			result := &gmodel.ServiceOtherInfo{
				Name:        item.Name,
				ServiceName: item.ServiceName,
				Remark:      item.Remark,
			}
			if len(result.Name) < 1 {
				result.Name = result.ServiceName
			}
			result.Status, _ = s.getStatus(result.ServiceName)

			results = append(results, result)
		}
	}

	ctx.Success(results)
}

func (s *Service) GetOthersDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, svcCatalogRoot, svcCatalogOther)
	function := catalog.AddFunction(method, uri, "获取服务列表")
	function.SetOutputDataExample([]*gmodel.ServiceOtherInfo{
		{
			Name:        "example",
			ServiceName: "other",
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}
