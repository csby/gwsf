package controller

import (
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
)

func (s *Monitor) GetDiskPartitionUsages(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.MonitorDiskArgument{}
	ctx.GetJson(argument)

	parts, err := gmonitor.DiskPartitions(argument.All)
	if err != nil {
		ctx.Error(gtype.ErrInternal.SetDetail(err))
		return
	}

	results := make([]*gmonitor.DiskUsage, 0)
	c := len(parts)
	for i := 0; i < c; i++ {
		p := parts[i]
		if p == nil {
			continue
		}

		path := p.MountPoint
		if len(path) < 1 {
			continue
		}

		u, e := gmonitor.StatDiskUsage(path)
		if e != nil {
			continue
		}
		if len(u.FsType) < 1 {
			u.FsType = p.FsType
		}

		results = append(results, u)
	}

	ctx.Success(results)
}

func (s *Monitor) GetDiskPartitionUsagesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogDisk)
	function := catalog.AddFunction(method, uri, "获取磁盘分区信息")
	function.SetInputJsonExample(&gmodel.MonitorDiskArgument{})
	function.SetOutputDataExample([]*gmonitor.DiskUsage{
		{
			Path:        "/",
			Total:       256 * 1024 * 1024 * 1024,
			Used:        128 * 1024 * 1024 * 1024,
			Free:        128 * 1024 * 1024 * 1024,
			UsedPercent: 50,
			TotalText:   "256GB",
			UsedText:    "128GB",
			FreeText:    "128GB",
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}
