package gmodel

import "github.com/csby/gwsf/gtype"

type MonitorCpuUsage struct {
	CpuName  string `json:"cpuName" note:"CPU名称"`
	MaxCount int    `json:"maxCount" note:"最大数量"`
	CurCount int    `json:"CurCount" note:"当前数量"`

	Percents MonitorCpuPercentCollection `json:"percents" note:"使用率"`
}

type MonitorCpuPercent struct {
	TimePoint int64 `json:"-" note:"时间点"`

	Time  gtype.DateTime `json:"time" note:"时间"`
	Usage float64        `json:"usage" note:"使用率"`
}

type MonitorCpuPercentCollection []*MonitorCpuPercent

func (s MonitorCpuPercentCollection) Len() int {
	return len(s)
}

func (s MonitorCpuPercentCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s MonitorCpuPercentCollection) Less(i, j int) bool {
	return s[i].TimePoint < s[j].TimePoint
}
