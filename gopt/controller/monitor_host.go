package controller

import (
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gtype"
	"time"
)

func (s *Monitor) GetHost(ctx gtype.Context, ps gtype.Params) {
	data := &gmonitor.Host{}
	err := data.Stat()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	ctx.Success(data)
}

func (s *Monitor) GetHostDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogHost)
	function := catalog.AddFunction(method, uri, "获取主机信息")
	function.SetNote("获取当前操作系统相关信息")
	function.SetOutputDataExample(&gmonitor.Host{
		ID:              "8f438ea2-c26b-401e-9f6b-19f2a0e4ee2e",
		Name:            "pc",
		BootTime:        gmonitor.DateTime(time.Now()),
		OS:              "linux",
		Platform:        "ubuntu",
		PlatformVersion: "18.04",
		KernelVersion:   "4.15.0-22-generic",
		CPU:             "Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz x2",
		Memory:          "4GB",
		TimeZone:        "GST+08",
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
