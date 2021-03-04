package gcfg

type Node struct {
	InstanceId  string `json:"-" note:"节点实例ID"`
	Enabled     bool   `json:"enabled" note:"是否启用"`
	CloudServer Cloud  `json:"cloudServer" note:"云服务器"`
	Cert        Crt    `json:"cert" note:"证书"`
}
