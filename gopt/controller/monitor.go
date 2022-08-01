package controller

import (
	"fmt"
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
)

const (
	monitorCatalogRoot    = "系统资源"
	monitorCatalogHost    = "主机"
	monitorCatalogDisk    = "磁盘"
	monitorCatalogNetwork = "网络"
	monitorCatalogCpu     = "处理器"
	monitorCatalogMemory  = "内存"
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
	function.SetOutputDataExample([]gmonitor.Listen{
		{
			Address:  "",
			Port:     22,
			Protocol: "tcp",
		},
	})
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrTokenInvalid)
}

func (s *Monitor) toSpeedText(v float64) string {
	kb := float64(1024)
	mb := 1024 * kb
	gb := 1024 * mb

	if v >= gb {
		return fmt.Sprintf("%.1fGbps", v/gb)
	} else if v >= mb {
		return fmt.Sprintf("%.1fMbps", v/mb)
	} else {
		return fmt.Sprintf("%.1fKbps", v/kb)
	}
}
