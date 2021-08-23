package gcfg

type Cloud struct {
	Address string `json:"address" note:"云服务器地址"`
	Port    int    `json:"port" note:"云服务器端口号"`
}

func (s *Cloud) CopyTo(target *Cloud) {
	if target == nil {
		return
	}

	target.Address = s.Address
	target.Port = s.Port
}
