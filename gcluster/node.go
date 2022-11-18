package gcluster

type Cluster struct {
	Index  uint64  `json:"index" note:"序号"`
	Enable bool    `json:"enable" note:"是否启用"`
	Nodes  []*Node `json:"nodes" note:"节点"`
}

type Node struct {
	Index   uint64     `json:"index" note:"序号"`
	Address string     `json:"address" note:"地址"`
	Port    int        `json:"port" note:"端口"`
	Status  NodeStatus `json:"status" note:"状态"`
}

type NodeStatus struct {
	In  bool `json:"in" note:"接收连接: true-已连接; false-未连接"`
	Out bool `json:"out" note:"发送连接: true-已连接; false-未连接"`
}
