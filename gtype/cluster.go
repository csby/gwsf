package gtype

type ClusterNodeStatus struct {
	Index uint64 `json:"index" note:"节点序号"`
	In    bool   `json:"in" note:"接收连接: true-已连接; false-未连接"`
	Out   bool   `json:"out" note:"发送连接: true-已连接; false-未连接"`
}
