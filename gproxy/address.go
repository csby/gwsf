package gproxy

import "sync"

type TargetAddress struct {
	sync.RWMutex

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
		if item.Alive == false {
			continue
		}

		if addr.Alive == false {
			addr = item
		} else if addr.Count > item.Count {
			addr = item
		}
	}

	return addr
}

func (s *TargetAddress) SetAddress(v string) {
	s.items = make([]*TargetAddressItem, 0)
	s.items = append(s.items, &TargetAddressItem{
		Addr:  v,
		Alive: false,
		Count: 0,
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
			Addr:  v,
			Alive: false,
			Count: 0,
		})
	}
}

func (s *TargetAddress) Items() []*TargetAddressItem {
	return s.items
}

func (s *TargetAddress) ResetCount() {
	s.Lock()
	defer s.Unlock()

	items := s.items
	c := len(items)
	for i := 1; i < c; i++ {
		item := items[i]
		if item == nil {
			continue
		}
		item.Count = 0
	}
}

type TargetAddressItem struct {
	sync.RWMutex

	Addr  string
	Alive bool
	Count int64
}

func (s *TargetAddressItem) IncreaseCount() {
	s.Lock()
	defer s.Unlock()

	s.Count += 1
}

func (s *TargetAddressItem) DecreaseCount() {
	s.Lock()
	defer s.Unlock()

	s.Count -= 1
}
