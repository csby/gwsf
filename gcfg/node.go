package gcfg

type Node struct {
	Enabled     bool  `json:"enabled" note:"是否启用"`
	CloudServer Cloud `json:"cloudServer" note:"云服务器"`
	Cert        Crt   `json:"cert" note:"证书"`
}
