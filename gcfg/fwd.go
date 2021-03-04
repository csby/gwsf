package gcfg

type Fwd struct {
	ID     string `json:"id" note:"标识ID"`
	Enable bool   `json:"enable" note:"是否启用"`

	ListenAddress string `json:"listenAddress" note:"监听地址"`
	ListenPort    int    `json:"listenPort" note:"监听端口"`

	TargetNodeID   string `json:"targetNodeId" note:"目标节点ID"`
	TargetNodeName string `json:"targetNodeName" note:"目标节点名称"`

	TargetAddress string `json:"targetAddress" note:"监听地址"`
	TargetPort    int    `json:"targetPort" note:"监听端口"`
}

func (s *Fwd) CopyTo(target *Fwd) {
	if target == nil {
		return
	}

	target.ID = s.ID
	target.Enable = s.Enable
	target.ListenAddress = s.ListenAddress
	target.ListenPort = s.ListenPort
	target.TargetNodeID = s.TargetNodeID
	target.TargetNodeName = s.TargetNodeName
	target.TargetAddress = s.TargetAddress
	target.TargetPort = s.TargetPort
}
