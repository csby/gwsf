package gcloud

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
	"github.com/gorilla/websocket"
	"sort"
	"strings"
	"sync"
)

type ForwardChannel struct {
	ID      string
	SrcConn *websocket.Conn
	DstConn chan *websocket.Conn
	Error   chan error

	StartTime  gtype.DateTime
	SourceNode *gtype.Node
	TargetNode *gtype.Node
	TargetAddr string
	TargetPort string
}

func (s *ForwardChannel) CopyTo(target *gtype.ForwardInfo) {
	if target == nil {
		return
	}

	target.ID = s.ID
	target.Time = s.StartTime
	target.SourceNode = s.SourceNode
	target.TargetNode = s.TargetNode
	target.TargetHost = fmt.Sprintf("%s:%s", s.TargetAddr, s.TargetPort)
}

func (s *ForwardChannel) match(filter *gtype.ForwardInfoFilter) bool {
	if filter == nil {
		return false
	}

	if len(filter.SourceIP) > 0 {
		if s.SourceNode == nil {
			return false
		}
		if !strings.Contains(s.SourceNode.IP, filter.SourceIP) {
			return false
		}
	}

	if len(filter.SourceName) > 0 {
		if s.SourceNode == nil {
			return false
		}
		if !strings.Contains(s.SourceNode.Name, filter.SourceName) {
			return false
		}
	}

	if len(filter.TargetIP) > 0 {
		if s.TargetNode == nil {
			return false
		}
		if !strings.Contains(s.TargetNode.IP, filter.TargetIP) {
			return false
		}
	}

	if len(filter.TargetName) > 0 {
		if s.TargetNode == nil {
			return false
		}
		if !strings.Contains(s.TargetNode.Name, filter.TargetName) {
			return false
		}
	}

	if len(filter.TargetHost) > 0 {
		if !strings.Contains(fmt.Sprintf("%s:%s", s.TargetAddr, s.TargetPort), filter.TargetHost) {
			return false
		}
	}

	return true
}

type ForwardChannelCollection struct {
	sync.RWMutex

	channels map[string]*ForwardChannel
}

func (s *ForwardChannelCollection) Add(ch *ForwardChannel) {
	s.Lock()
	defer s.Unlock()
	if ch == nil {
		return
	}
	id := ch.ID
	if len(id) < 1 {
		return
	}

	s.channels[id] = ch
}

func (s *ForwardChannelCollection) Get(id string) *ForwardChannel {
	s.Lock()
	defer s.Unlock()

	if len(id) < 1 {
		return nil
	}

	v, ok := s.channels[id]
	if !ok {
		return nil
	}

	return v
}

func (s *ForwardChannelCollection) Del(id string) {
	s.Lock()
	defer s.Unlock()

	if len(id) < 1 {
		return
	}

	_, ok := s.channels[id]
	if ok {
		delete(s.channels, id)
	}
}

func (s *ForwardChannelCollection) Lst(filter *gtype.ForwardInfoFilter) gtype.ForwardInfoArray {
	s.RLock()
	defer s.RUnlock()

	items := make(gtype.ForwardInfoArray, 0)
	for k, v := range s.channels {
		if len(k) < 1 {
			continue
		}

		if filter != nil {
			if !v.match(filter) {
				continue
			}
		}

		item := &gtype.ForwardInfo{}
		v.CopyTo(item)
		items = append(items, item)
	}

	sort.Sort(items)
	return items
}
