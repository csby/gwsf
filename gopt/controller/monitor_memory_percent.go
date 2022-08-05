package controller

import "sync"

type NetworkMemoryPercent struct {
	TimePoint    int64   // 时间点
	UsagePercent float64 // 使用率
	Total        uint64  // 总共
	Used         uint64  // 已使用
}

type NetworkMemoryUsage struct {
	sync.RWMutex

	Count int // 最大数量

	percents []*NetworkMemoryPercent
}

func (s *NetworkMemoryUsage) Percents() []*NetworkMemoryPercent {
	return s.percents
}

func (s *NetworkMemoryUsage) Reset(count int) {
	s.Lock()
	defer s.Unlock()

	s.Count = count
}

func (s *NetworkMemoryUsage) AddValue(t int64, total, used uint64, usage float64) *NetworkMemoryPercent {
	s.Lock()
	defer s.Unlock()

	if s.percents == nil {
		s.percents = make([]*NetworkMemoryPercent, 0)
	}

	if s.Count < 1 {
		return nil
	}

	var percent *NetworkMemoryPercent
	c := len(s.percents)
	if c < s.Count {
		percent = &NetworkMemoryPercent{}
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
	percent.Total = total
	percent.Used = used
	percent.UsagePercent = usage

	return percent
}
