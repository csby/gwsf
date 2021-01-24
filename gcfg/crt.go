package gcfg

type Crt struct {
	Ca     CrtCa  `json:"ca" note:"CA证书"`
	Server CrtPfx `json:"server" note:"服务器证书"`
}
