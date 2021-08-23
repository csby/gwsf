package gcfg

type FwdId struct {
	ID string `json:"id" required:"true" note:"标识ID"`
}

type FwdEnable struct {
	Enable bool `json:"enable" note:"是否启用"`
}

type FwdContent struct {
	FwdEnable

	ListenAddress string `json:"listenAddress" note:"监听地址"`
	ListenPort    int    `json:"listenPort" note:"监听端口"`

	TargetNodeID   string `json:"targetNodeId" note:"目标结点ID"`
	TargetNodeName string `json:"targetNodeName" note:"目标结点名称"`

	TargetAddress string `json:"targetAddress" note:"目标地址"`
	TargetPort    int    `json:"targetPort" note:"目标端口"`
}

type Fwd struct {
	FwdId
	FwdContent
}

type FwdState struct {
	IsRunning bool   `json:"running" note:"状态: true-运行中; false-已停止"`
	LastError string `json:"error" note:"错误信息"`
}

type FwdInfo struct {
	Fwd
	FwdState
}

func (s *FwdContent) CopyTo(target *Fwd) {
	if target == nil {
		return
	}

	target.Enable = s.Enable
	target.ListenAddress = s.ListenAddress
	target.ListenPort = s.ListenPort
	target.TargetNodeID = s.TargetNodeID
	target.TargetNodeName = s.TargetNodeName
	target.TargetAddress = s.TargetAddress
	target.TargetPort = s.TargetPort
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

func (s *FwdInfo) CopyFrom(source *Fwd) {
	if source == nil {
		return
	}
	source.CopyTo(&s.Fwd)
}
