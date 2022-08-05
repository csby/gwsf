package gmodel

import "github.com/csby/gwsf/gtype"

type MonitorNetworkIOArgument struct {
	Name string `json:"name" note:"网卡名称"`
}

type MonitorNetworkIO struct {
	Name     string `json:"name" note:"网卡名称"`
	MaxCount int    `json:"maxCount" note:"流量最大数量"`
	CurCount int    `json:"CurCount" note:"流量当前数量"`

	Flows MonitorNetworkIOThroughputCollection `json:"flows" note:"流量"`
}

type MonitorNetworkIOThroughput struct {
	TimePoint int64 `json:"-" note:"时间点"`

	Time           gtype.DateTime `json:"time" note:"时间"`
	BytesSpeedSent uint64         `json:"bytesSpeedSent" note:"发送字节速率, 单位: 字节/秒"`
	BytesSpeedRecv uint64         `json:"bytesSpeedRecv" note:"接收字节速率, 单位: 字节/秒"`

	BytesSpeedSentText string `json:"bytesSpeedSentText" note:"发送字节速率文本信息"`
	BytesSpeedRecvText string `json:"bytesSpeedRecvText" note:"接收字节速率文本信息"`
}

type MonitorNetworkIOThroughputCollection []*MonitorNetworkIOThroughput

func (s MonitorNetworkIOThroughputCollection) Len() int {
	return len(s)
}

func (s MonitorNetworkIOThroughputCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s MonitorNetworkIOThroughputCollection) Less(i, j int) bool {
	return s[i].TimePoint < s[j].TimePoint
}

type MonitorNetworkIOThroughputArgument struct {
	Name string `json:"name" note:"网卡名称"`

	Flow MonitorNetworkIOThroughput `json:"flow" note:"吞吐量"`
}
