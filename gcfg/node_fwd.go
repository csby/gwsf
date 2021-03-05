package gcfg

import "github.com/csby/gwsf/gtype"

type NodeFwd struct {
	Enable bool `json:"enable" note:"是否启用"`

	Tcp []*Fwd `json:"tcp" note:"TCP转发列表"`
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
