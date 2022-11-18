package gcfg

type ClusterInstance struct {
	Index   uint64 `json:"index" note:"序号，有效值：1～9"`
	Address string `json:"address" note:"地址"`
	Port    int    `json:"port" note:"端口"`
	Secure  bool   `json:"secure" note:"是否启用安全连接"`
	Ca      CrtCa  `json:"ca" note:"CA证书"`
	Crt     CrtPfx `json:"crt" note:"实例证书"`
}
