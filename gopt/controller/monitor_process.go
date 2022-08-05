package controller

import (
	"fmt"
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
)

func (s *Monitor) GetProcessInfo(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.ProcessID{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput, err)
		return
	}
	if argument.Pid < 1 {
		ctx.Error(gtype.ErrInput, fmt.Sprintf("无效的进程ID: %d", argument.Pid))
		return
	}
	info, err := gmonitor.GetProcessInfo(argument.Pid)
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}

	ctx.Success(info)
}

func (s *Monitor) GetProcessInfoDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogProcess)
	function := catalog.AddFunction(method, uri, "获取进程信息")
	function.SetNote("获取指定进程相关信息")
	function.SetInputJsonExample(&gmodel.ProcessID{})
	function.SetOutputDataExample(&gmonitor.Process{})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInput)
}
