package gproxy

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"sync"
)

type TargetAddress struct {
	sync.RWMutex

	SourceId string
	TargetId string

	AliveChanged func(item *TargetAddressItem)
	CountChanged func(item *TargetAddressItem, increase bool)

	items []*TargetAddressItem
}

func (s *TargetAddress) GetAddress() *TargetAddressItem {
	items := s.items
	c := len(items)
	if c < 1 {
		return nil
	}

	addr := items[0]
	if c == 1 {
		return addr
	}

	s.Lock()
	defer s.Unlock()
	for i := 1; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}
		if item.IstAlive() == false {
			continue
		}

		if addr.IstAlive() == false {
			addr = item
		} else if addr.Count() > item.Count() {
			addr = item
		}
	}

	return addr
}

func (s *TargetAddress) SetAddress(v string) {
	s.items = make([]*TargetAddressItem, 0)
	s.items = append(s.items, &TargetAddressItem{
		SourceId:     s.SourceId,
		TargetId:     s.TargetId,
		AddrId:       gtype.ToMd5(fmt.Sprintf("%s-%s-%s", s.SourceId, s.TargetId, v)),
		Addr:         v,
		alive:        false,
		count:        0,
		aliveChanged: s.AliveChanged,
		countChanged: s.CountChanged,
	})
}

func (s *TargetAddress) AddAddress(vs []string) {
	if s.items == nil {
		s.items = make([]*TargetAddressItem, 0)
	}

	c := len(vs)
	for i := 0; i < c; i++ {
		v := vs[i]
		if len(v) < 1 {
			continue
		}

		s.items = append(s.items, &TargetAddressItem{
			SourceId:     s.SourceId,
			TargetId:     s.TargetId,
			AddrId:       gtype.ToMd5(fmt.Sprintf("%s-%s-%s", s.SourceId, s.TargetId, v)),
			Addr:         v,
			alive:        false,
			count:        0,
			aliveChanged: s.AliveChanged,
			countChanged: s.CountChanged,
		})
	}
}

func (s *TargetAddress) Items() []*TargetAddressItem {
	return s.items
}

type TargetAddressItem struct {
	sync.RWMutex

	SourceId string
	TargetId string
	AddrId   string
	Addr     string

	alive bool
	count int64

	aliveChanged func(item *TargetAddressItem)
	countChanged func(item *TargetAddressItem, increase bool)
}

func (s *TargetAddressItem) SetAlive(v bool) {
	if s.alive == v {
		return
	}
	s.alive = v

	if v {
		s.setCount(0)
	}

	go s.fireAliveChanged()
}

func (s *TargetAddressItem) IstAlive() bool {
	return s.alive
}

func (s *TargetAddressItem) Count() int64 {
	return s.count
}

func (s *TargetAddressItem) IncreaseCount() {
	s.Lock()
	defer s.Unlock()

	s.setCount(s.count + 1)

	go s.fireCountChanged(true)
}

func (s *TargetAddressItem) DecreaseCount() {
	s.Lock()
	defer s.Unlock()

	s.setCount(s.count - 1)

	go s.fireCountChanged(false)
}

func (s *TargetAddressItem) setCount(v int64) {
	if v < 0 {
		v = 0
	}

	if s.count == v {
		return
	}
	s.count = v
}

func (s *TargetAddressItem) fireAliveChanged() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	if s.aliveChanged != nil {
		s.aliveChanged(s)
	}
}

func (s *TargetAddressItem) fireCountChanged(increase bool) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	if s.countChanged != nil {
		s.countChanged(s, increase)
	}
}
