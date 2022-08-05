package controller

import (
	"fmt"
	"github.com/csby/gmonitor"
	"github.com/csby/gwsf/gmodel"
	"github.com/csby/gwsf/gtype"
	"sort"
	"time"
)

func (s *Monitor) GetNetworkInterfaces(ctx gtype.Context, ps gtype.Params) {
	data, err := gmonitor.Interfaces()
	if err != nil {
		ctx.Error(gtype.ErrInternal, err)
		return
	}
	ctx.Success(data)
}

func (s *Monitor) GetNetworkInterfacesDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogNetwork)
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

	c := len(data)
	for i := 0; i < c; i++ {
		item := data[i]
		if item == nil {
			continue
		}
		pid := item.PId
		if pid < 1 {
			continue
		}
	}

	ctx.Success(data)
}

func (s *Monitor) GetNetworkListenPortsDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogNetwork)
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

func (s *Monitor) GetNetworkThroughput(ctx gtype.Context, ps gtype.Params) {
	argument := &gmodel.MonitorNetworkIOArgument{}
	err := ctx.GetJson(argument)
	if err != nil {
		ctx.Error(gtype.ErrInput.SetDetail(err))
		return
	}
	if len(argument.Name) < 1 {
		ctx.Error(gtype.ErrInput.SetDetail("网卡名称(name)为空"))
		return
	}

	face := s.faces.GetInterface(argument.Name)
	if face == nil {
		ctx.Error(gtype.ErrInput.SetDetail(fmt.Sprintf("网卡名称(%s)不存在", argument.Name)))
		return
	}

	results := make(gmodel.MonitorNetworkIOThroughputCollection, 0)
	items := face.Counters()
	c := len(items)
	for i := 0; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}

		interval := uint64(item.TimeInterval)
		if interval < 1 {
			continue
		}
		result := &gmodel.MonitorNetworkIOThroughput{
			TimePoint:      item.TimePoint,
			BytesSpeedSent: item.BytesSent / interval,
			BytesSpeedRecv: item.BytesRecv / interval,
		}
		result.Time = gtype.DateTime(time.Unix(item.TimePoint, 0))
		result.BytesSpeedSentText = s.toSpeedText(float64(item.BytesSent) / float64(interval))
		result.BytesSpeedRecvText = s.toSpeedText(float64(item.BytesRecv) / float64(interval))

		results = append(results, result)
	}

	sort.Sort(results)
	ctx.Success(&gmodel.MonitorNetworkIO{
		Name:     argument.Name,
		MaxCount: s.faces.MaxCounter,
		CurCount: len(results),
		Flows:    results,
	})
}

func (s *Monitor) GetNetworkThroughputDoc(doc gtype.Doc, method string, uri gtype.Uri) {
	catalog := s.createCatalog(doc, monitorCatalogRoot, monitorCatalogNetwork)
	function := catalog.AddFunction(method, uri, "获取网卡吞吐量")
	function.SetInputJsonExample(&gmodel.MonitorNetworkIOArgument{
		Name: "eth0",
	})
	function.SetOutputDataExample(&gmodel.MonitorNetworkIO{
		Name:     "eth0",
		MaxCount: 60,
		CurCount: 1,
		Flows: []*gmodel.MonitorNetworkIOThroughput{
			{
				Time:               gtype.DateTime(time.Now()),
				BytesSpeedSent:     3 * 1024,
				BytesSpeedRecv:     5 * 1024,
				BytesSpeedSentText: s.toSpeedText(float64(3 * 1024)),
				BytesSpeedRecvText: s.toSpeedText(float64(5 * 1024)),
			},
		},
	})
	function.AddOutputError(gtype.ErrTokenEmpty)
	function.AddOutputError(gtype.ErrTokenInvalid)
	function.AddOutputError(gtype.ErrInternal)
	function.AddOutputError(gtype.ErrInput)
}

func (s *Monitor) doStatNetworkIO(interval time.Duration) {
	for {
		time.Sleep(interval)

		items, err := gmonitor.StatNetworkIOs()
		if err != nil {
			continue
		}

		t := time.Now().Unix()
		c := len(items)
		for i := 0; i < c; i++ {
			item := items[i]
			if item == nil {
				continue
			}

			counter := s.faces.AddIOCounter(t, item)
			if counter != nil {
				s.sentNetworkThroughput(item.Name, counter)
			}
		}
	}
}

func (s *Monitor) sentNetworkThroughput(name string, item *NetworkIOCounter) {
	if item == nil {
		return
	}
	interval := uint64(item.TimeInterval)
	if interval < 1 {
		return
	}

	argument := &gmodel.MonitorNetworkIOThroughputArgument{
		Name: name,
		Flow: gmodel.MonitorNetworkIOThroughput{
			TimePoint:      item.TimePoint,
			BytesSpeedSent: item.BytesSent / interval,
			BytesSpeedRecv: item.BytesRecv / interval,
		},
	}
	argument.Flow.Time = gtype.DateTime(time.Unix(item.TimePoint, 0))
	argument.Flow.BytesSpeedSentText = s.toSpeedText(float64(item.BytesSent) / float64(interval))
	argument.Flow.BytesSpeedRecvText = s.toSpeedText(float64(item.BytesRecv) / float64(interval))

	go s.writeWebSocketMessage("", gtype.WSNetworkThroughput, argument)
}
