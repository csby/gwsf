package gcfg

import (
	"fmt"
	"github.com/csby/gwsf/gtype"
)

type NodeFwd struct {
	Enable bool `json:"enable" note:"是否启用"`

	Tcp []*Fwd `json:"tcp" note:"TCP转发列表"`
	Udp []*Fwd `json:"udp" note:"UDP"`
}

func (s *NodeFwd) InitId() {
	c := len(s.Tcp)
	for i := 0; i < c; i++ {
		item := s.Tcp[i]
		if item == nil {
			continue
		}

		if len(item.ID) > 0 {
			continue
		}

		item.ID = gtype.NewGuid()
	}
}

func (s *NodeFwd) GetTcpFwd(id string) *Fwd {
	count := len(s.Tcp)
	for index := 0; index < count; index++ {
		item := s.Tcp[index]
		if item == nil {
			continue
		}
		if id == item.ID {
			return item
		}
	}

	return nil
}

func (s *NodeFwd) GetUdpFwd(id string) *Fwd {
	count := len(s.Udp)
	for index := 0; index < count; index++ {
		item := s.Udp[index]
		if item == nil {
			continue
		}
		if id == item.ID {
			return item
		}
	}

	return nil
}

func (s *NodeFwd) GetTcpFwdId(listenAddress string, listenPort int) (string, bool) {
	addr := fmt.Sprintf("%s:%d", listenAddress, listenPort)
	count := len(s.Tcp)
	for index := 0; index < count; index++ {
		item := s.Tcp[index]
		if item == nil {
			continue
		}
		if addr == fmt.Sprintf("%s:%d", item.ListenAddress, item.ListenPort) {
			return item.ID, true
		}
	}

	return "", false
}

func (s *NodeFwd) GetUdpFwdId(listenAddress string, listenPort int) (string, bool) {
	addr := fmt.Sprintf("%s:%d", listenAddress, listenPort)
	count := len(s.Udp)
	for index := 0; index < count; index++ {
		item := s.Udp[index]
		if item == nil {
			continue
		}
		if addr == fmt.Sprintf("%s:%d", item.ListenAddress, item.ListenPort) {
			return item.ID, true
		}
	}

	return "", false
}
