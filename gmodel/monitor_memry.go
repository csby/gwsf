package gmodel

import "github.com/csby/gwsf/gtype"

type MonitorMemoryPercent struct {
	TimePoint int64 `json:"-" note:"时间点"`

	Time  gtype.DateTime `json:"time" note:"时间"`
	Total uint64         `json:"total" note:"内存过总数"`
	Used  uint64         `json:"used" note:"已使用数"`
	Usage float64        `json:"usage" note:"已使用率"`
}

type MonitorMemoryPercentCollection []*MonitorMemoryPercent

func (s MonitorMemoryPercentCollection) Len() int {
	return len(s)
}

func (s MonitorMemoryPercentCollection) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s MonitorMemoryPercentCollection) Less(i, j int) bool {
	return s[i].TimePoint < s[j].TimePoint
}

type MonitorMemoryUsage struct {
	MaxCount int `json:"maxCount" note:"最大数量"`
	CurCount int `json:"CurCount" note:"当前数量"`

	Percents MonitorMemoryPercentCollection `json:"percents" note:"使用率"`
}
