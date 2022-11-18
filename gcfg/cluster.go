package gcfg

type Cluster struct {
	Index     uint64            `json:"index" note:"索引序号，有效值：1～9"`
	Enable    bool              `json:"enable" note:"是否加入集群"`
	Instances []ClusterInstance `json:"instances" note:"集群实例"`
}
