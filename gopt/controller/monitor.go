package controller

import (
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"time"
)

const (
	monitorCatalogRoot    = "系统资源"
	monitorCatalogHost    = "主机"
	monitorCatalogDisk    = "磁盘"
	monitorCatalogNetwork = "网络"
	monitorCatalogCpu     = "处理器"
	monitorCatalogMemory  = "内存"
	monitorCatalogProcess = "进程"
)

type Monitor struct {
	controller

	faces    *NetworkInterfaceCollection
	cpuUsage *NetworkCpuUsage
	cupName  string
	memUsage *NetworkMemoryUsage
}

func NewMonitor(log gtype.Log, cfg *gcfg.Config, wsc gtype.SocketChannelCollection) *Monitor {
	inst := &Monitor{}
	inst.SetLog(log)
	inst.cfg = cfg
	inst.wsChannels = wsc

	maxCount := 30
	inst.faces = &NetworkInterfaceCollection{
		MaxCounter: maxCount,
	}
	inst.cpuUsage = &NetworkCpuUsage{
		Count: maxCount,
	}
	inst.memUsage = &NetworkMemoryUsage{
		Count: maxCount,
	}

	interval := time.Second
	go inst.doStatNetworkIO(interval)
	go inst.doStatCpuUsage(interval)
	go inst.doStatMemoryUsage(10 * interval)

	return inst
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
