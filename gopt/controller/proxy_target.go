package controller

type ProxyTargetItem interface {
	SourceId() string
	TargetId() string
	IsAlive() bool
	Count() int64
	SetAlive(v bool)
	IncreaseCount() int64
	DecreaseCount() int64
}

type ProxyTargetCollection struct {
	items map[string]ProxyTargetItem
}

func (s *ProxyTargetCollection) AddItem(id string, item ProxyTargetItem) {
	if item == nil {
		return
	}

	s.items[id] = item
}

func (s *ProxyTargetCollection) GetItem(id string) ProxyTargetItem {
	if s.items == nil {
		return nil
	}

	item, ok := s.items[id]
	if ok {
		return item
	}

	return nil
}

func (s *ProxyTargetCollection) Stop() {
	if s.items == nil {
		return
	}

	for _, item := range s.items {
		if item == nil {
			continue
		}

		item.SetAlive(false)
	}
}

func (s *ProxyTargetCollection) SetAlive(id string, v bool) {
	if s.items == nil {
		return
	}

	item, ok := s.items[id]
	if !ok {
		return
	}
	if item == nil {
		return
	}

	item.SetAlive(v)
}

type ProxyTargetEntry struct {
	SourceId string `json:"sourceId" note:"服务ID"`
	TargetId string `json:"targetId" note:"目标ID"`
	AddrId   string `json:"addrId" note:"地址ID "`
	Alive    bool   `json:"alive" note:"是否在线"`
	Count    int64  `json:"count" note:"连接数量"`
}
