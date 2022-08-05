package controller

import (
	"github.com/csby/gmonitor"
	"sync"
)

type NetworkIOCounter struct {
	TimePoint    int64  // 时间点
	TimeInterval int64  // 时间间隔
	BytesSent    uint64 // 发送字节
	BytesRecv    uint64 // 接收字节
	PacketsSent  uint64 // 发送数据包
	PacketsRecv  uint64 // 接收数据包
}

type NetworkInterface struct {
	sync.RWMutex

	Name  string // 网卡名称
	Count int    // 最大数量

	last     *NetworkIOCounter // 最后一次
	counters []*NetworkIOCounter
}

func (s *NetworkInterface) AddCounter(time int64, bytesSent, bytesRecv, packetsSent, packetsRecv uint64) *NetworkIOCounter {
	s.Lock()
	defer s.Unlock()

	if s.last == nil {
		s.last = &NetworkIOCounter{}
		s.last.TimePoint = time
		s.last.BytesSent = bytesSent
		s.last.BytesRecv = bytesRecv
		s.last.PacketsSent = packetsSent
		s.last.PacketsRecv = packetsRecv
		return nil
	}
	if s.Count < 1 {
		return nil
	}

	if s.counters == nil {
		s.counters = make([]*NetworkIOCounter, 0)
	}

	var counter *NetworkIOCounter
	c := len(s.counters)
	if c < s.Count {
		counter = &NetworkIOCounter{}
		s.counters = append(s.counters, counter)
	} else {
		counter = s.counters[0]
		for i := 1; i < c; i++ {
			item := s.counters[i]
			if counter.TimePoint > item.TimePoint {
				counter = item
			}
		}
	}

	counter.TimePoint = time
	counter.TimeInterval = time - s.last.TimePoint
	counter.BytesSent = bytesSent - s.last.BytesSent
	counter.BytesRecv = bytesRecv - s.last.BytesRecv
	counter.PacketsSent = packetsSent - s.last.PacketsSent
	counter.PacketsRecv = packetsRecv - s.last.PacketsRecv

	s.last.TimePoint = time
	s.last.BytesSent = bytesSent
	s.last.BytesRecv = bytesRecv
	s.last.PacketsSent = packetsSent
	s.last.PacketsRecv = packetsRecv

	return counter
}

func (s *NetworkInterface) Counters() []*NetworkIOCounter {
	return s.counters
}

func (s *NetworkInterface) Reset(count int) {
	s.Lock()
	defer s.Unlock()

	s.Count = count
	s.last = nil
	s.counters = make([]*NetworkIOCounter, 0)
}

type NetworkInterfaceCollection struct {
	sync.RWMutex

	items map[string]*NetworkInterface

	MaxCounter int
}

func (s *NetworkInterfaceCollection) AddIOCounter(t int64, v *gmonitor.NetworkIO) *NetworkIOCounter {
	if v == nil {
		return nil
	}
	if len(v.Name) < 1 {
		return nil
	}

	face := s.getInterface(v.Name)
	if nil == face {
		return nil
	}

	return face.AddCounter(t, v.BytesSent, v.BytesRecv, v.PacketsSent, v.PacketsRecv)
}

func (s *NetworkInterfaceCollection) Reset() {
	s.Lock()
	defer s.Unlock()

	if len(s.items) < 1 {
		return
	}

	for _, v := range s.items {
		if v == nil {
			continue
		}
		v.Reset(s.MaxCounter)
	}
}

func (s *NetworkInterfaceCollection) GetInterface(name string) *NetworkInterface {
	if len(name) < 1 {
		return nil
	}
	if s.items == nil {
		return nil
	}
	v, ok := s.items[name]
	if !ok {
		return nil
	}

	return v
}

func (s *NetworkInterfaceCollection) getInterface(name string) *NetworkInterface {
	s.Lock()
	defer s.Unlock()

	if s.items == nil {
		s.items = make(map[string]*NetworkInterface)
	}
	v, ok := s.items[name]
	if !ok {
		v := &NetworkInterface{
			Name:  name,
			Count: s.MaxCounter,
		}
		s.items[name] = v
	}

	return v
}
