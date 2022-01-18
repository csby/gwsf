package gcfg

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
)

type NodeFwd struct {
	Enable bool `json:"enable" note:"是否启用"`

	Items []*Fwd `json:"items" note:"转发列表"`
}

func (s *NodeFwd) InitId() {
	c := len(s.Items)
	for i := 0; i < c; i++ {
		item := s.Items[i]
		if item == nil {
			continue
		}

		if len(item.ID) > 0 {
			continue
		}

		item.ID = gtype.NewGuid()
	}
}

func (s *NodeFwd) GetItem(id string) *Fwd {
	count := len(s.Items)
	for index := 0; index < count; index++ {
		item := s.Items[index]
		if item == nil {
			continue
		}
		if id == item.ID {
			return item
		}
	}

	return nil
}

func (s *NodeFwd) GetItemId(listenAddress string, listenPort int) (string, bool) {
	addr := fmt.Sprintf("%s:%d", listenAddress, listenPort)
	count := len(s.Items)
	for index := 0; index < count; index++ {
		item := s.Items[index]
		if item == nil {
			continue
		}
		if addr == fmt.Sprintf("%s:%d", item.ListenAddress, item.ListenPort) {
			return item.ID, true
		}
	}

	return "", false
}
