package gproxy

import (
	"github.com/csby/gwsf/gtype"
	"sort"
	"strings"
	"sync"
)

type Link struct {
	Id         string         `json:"id" note:"标识ID"`
	Time       gtype.DateTime `json:"time" note:"时间"`
	ListenAddr string         `json:"listenAddr" note:"监听地址"`
	Domain     string         `json:"domain" note:"域名"`
	SourceAddr string         `json:"sourceAddr" note:"传入地址"`
	TargetAddr string         `json:"targetAddr" note:"目标地址"`
}

type LinkFilter struct {
	ListenAddr string `json:"listenAddr" note:"监听地址"`
	Domain     string `json:"domain" note:"域名"`
	TargetAddr string `json:"targetAddr" note:"目标地址"`
}

func (s *LinkFilter) match(link *Link) bool {
	if link == nil {
		return false
	}

	if len(s.ListenAddr) > 0 {
		if !strings.Contains(link.ListenAddr, s.ListenAddr) {
			return false
		}
	}

	if len(s.Domain) > 0 {
		if !strings.Contains(link.Domain, s.Domain) {
			return false
		}
	}

	if len(s.TargetAddr) > 0 {
		if !strings.Contains(link.TargetAddr, s.TargetAddr) {
			return false
		}
	}

	return true
}

type LinkCollection interface {
	Add(item *Link)
	Del(id string) bool
	Lst(filter *LinkFilter) []*Link
}

func NewLinkCollection() LinkCollection {
	return &linkCollection{
		items: make(map[string]*Link),
	}
}

type linkCollection struct {
	sync.RWMutex

	items map[string]*Link
}

func (s *linkCollection) Add(item *Link) {
	if nil == item {
		return
	}
	if len(item.Id) < 1 {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.items[item.Id] = item
}

func (s *linkCollection) Del(id string) bool {
	s.Lock()
	defer s.Unlock()

	_, ok := s.items[id]
	if ok {
		delete(s.items, id)
	}

	return ok
}

func (s *linkCollection) Lst(filter *LinkFilter) []*Link {
	s.RLock()
	defer s.RUnlock()

	items := make(linkArray, 0)
	for k, v := range s.items {
		if len(k) < 1 {
			continue
		}

		if filter != nil {
			if !filter.match(v) {
				continue
			}
		}

		items = append(items, v)
	}

	sort.Sort(items)
	return items
}

type linkArray []*Link

func (s linkArray) Len() int {
	return len(s)
}

func (s linkArray) Less(i, j int) bool {
	if s[i].Time.After(s[j].Time) {
		return true
	}

	return false
}

func (s linkArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
