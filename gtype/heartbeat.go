package gtype

import "sync"

type Heartbeat struct {
	Id       string `json:"id" note:"标识"`
	Protocol string `json:"protocol" note:"协议"`
	Host     string `json:"host" note:"主机"`
	Remote   string `json:"remote" note:"远程地址"`
}

type HeartbeatArray struct {
	sync.RWMutex

	Items []*Heartbeat
}

func (s *HeartbeatArray) Add(item *Heartbeat) {
	if item == nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.Items = append(s.Items, item)
}

func (s *HeartbeatArray) Del(id string) {
	s.Lock()
	defer s.Unlock()

	items := make([]*Heartbeat, 0)
	c := len(s.Items)
	for i := 0; i < c; i++ {
		item := s.Items[i]
		if item == nil {
			continue
		}
		if item.Id == id {
			continue
		}

		items = append(items, item)
	}

	s.Items = items
}
