package controller

import (
	"github.com/csby/gmonitor"
	"sync"
)

type NetworkCpuPercent struct {
	TimePoint    int64   // 时间点
	UsagePercent float64 // 使用率
}

type NetworkCpuUsage struct {
	sync.RWMutex

	Count int // 最大数量

	lastAll  float64
	lastBusy float64
	percents []*NetworkCpuPercent
}

func (s *NetworkCpuUsage) Percents() []*NetworkCpuPercent {
	return s.percents
}

func (s *NetworkCpuUsage) Reset(count int) {
	s.Lock()
	defer s.Unlock()

	s.Count = count
	s.percents = nil
}

func (s *NetworkCpuUsage) AddTime(t int64, all, busy float64) *NetworkCpuPercent {
	s.Lock()
	defer s.Unlock()

	if s.percents == nil {
		s.percents = make([]*NetworkCpuPercent, 0)
		s.lastAll = all
		s.lastBusy = busy
		return nil
	}

	if s.Count < 1 {
		return nil
	}

	var percent *NetworkCpuPercent
	c := len(s.percents)
	if c < s.Count {
		percent = &NetworkCpuPercent{}
		s.percents = append(s.percents, percent)
	} else {
		percent = s.percents[0]
		for i := 1; i < c; i++ {
			item := s.percents[i]
			if percent.TimePoint > item.TimePoint {
				percent = item
			}
		}
	}

	percent.TimePoint = t
	percent.UsagePercent = gmonitor.ToCpuPercent(s.lastAll, s.lastBusy, all, busy)
	s.lastAll = all
	s.lastBusy = busy

	return percent
}
