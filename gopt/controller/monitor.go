package controller

import (
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"time"
)

type Monitor struct {
	controller
}

func NewMonitor(log gtype.Log, cfg *gcfg.Config) *Monitor {
	instance := &Monitor{}
	instance.SetLog(log)
	instance.cfg = cfg

	return instance
}

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
	catalog := s.createCatalog(doc, "系统信息")
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

func (s *Monitor) GetNetworkInterfaces(ctx gtype.Context, ps gtype.Params) {
	data, err := gmonitor.Interfaces()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	ctx.Success(data)
}

func (s *Monitor) GetNetworkInterfacesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "系统信息")
	function := catalog.AddFunction(method, uri, "获取网卡信息")
	function.SetNote("获取主机网卡相关信息")
	function.SetOutputDataExample([]gmonitor.Interface{
		{
			Name:       "本地连接",
			MTU:        1500,
			MacAddress: "00:16:5d:13:b9:70",
			IPAddress: []string{
				"fe80::b1d0:ff08:1f6f:3e0b/64",
				"192.168.1.1/24",
			},
			Flags: []string{
				"up",
				"broadcast",
				"multicast",
			},
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Monitor) GetNetworkListenPorts(ctx gtype.Context, ps gtype.Params) {
	data := gmonitor.ListenPorts()
	ctx.Success(data)
}

func (s *Monitor) GetNetworkListenPortsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, "系统信息")
	function := catalog.AddFunction(method, uri, "获取监听端口")
	function.SetNote("获取主机正在监听端口信息")
	function.SetOutputDataExample([]gmonitor.Interface{
		{
			Name:       "本地连接",
			MTU:        1500,
			MacAddress: "00:16:5d:13:b9:70",
			IPAddress: []string{
				"fe80::b1d0:ff08:1f6f:3e0b/64",
				"192.168.1.1/24",
			},
			Flags: []string{
				"up",
				"broadcast",
				"multicast",
			},
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}
