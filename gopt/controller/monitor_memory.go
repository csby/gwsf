package controller

import (
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"sort"
	"time"
)

func (s *Monitor) GetMemoryUsage(ctx gtype.Context, ps gtype.Params) {
	results := make(gmodel.MonitorMemoryPercentCollection, 0)
	items := s.memUsage.Percents()
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}

		result := &gmodel.MonitorMemoryPercent{
			TimePoint: item.TimePoint,
			Usage:     item.UsagePercent,
			Total:     item.Total,
			Used:      item.Used,
		}
		result.Time = gtype.DateTime(time.Unix(item.TimePoint, 0))

		results = append(results, result)
	}

	sort.Sort(results)
	ctx.Success(&gmodel.MonitorMemoryUsage{
		MaxCount: s.faces.MaxCounter,
		CurCount: len(results),
		Percents: results,
	})
}

func (s *Monitor) GetMemoryUsageDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogMemory)
	function := catalog.AddFunction(method, uri, "获取内存使用率")
	function.SetOutputDataExample(&gmodel.MonitorMemoryUsage{
		MaxCount: 60,
		CurCount: 1,
		Percents: []*gmodel.MonitorMemoryPercent{
			{
				Time:  gtype.DateTime(time.Now()),
				Usage: 50,
				Total: 1024 * 1024 * 1024 * 16,
				Used:  1024 * 1024 * 1024 * 8,
			},
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Monitor) doStatMemoryUsage(interval time.Duration) {
	memory := &gmonitor.Memory{}
	for {
		time.Sleep(interval)

		err := memory.Stat()
		if err != nil {
			continue
		}

		t := time.Now().Unix()
		percent := s.memUsage.AddValue(t, memory.Total, memory.Used, memory.UsedPercent)
		if percent != nil {
			s.sentMemoryUsage(percent)
		}
	}
}

func (s *Monitor) sentMemoryUsage(item *NetworkMemoryPercent) {
	if item == nil {
		return
	}

	argument := &gmodel.MonitorMemoryPercent{
		Time:  gtype.DateTime(time.Unix(item.TimePoint, 0)),
		Usage: item.UsagePercent,
		Total: item.Total,
		Used:  item.Used,
	}

	go s.writeWebSocketMessage("", gtype.WSMemUsage, argument)
}
