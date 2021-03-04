package gcfg

type NodeFwd struct {
	Enable bool `json:"enable" note:"是否启用"`

	Tcp []*Fwd `json:"tcp" note:"TCP转发列表"`
}
