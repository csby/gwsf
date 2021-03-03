package gcfg

type Cloud struct {
	Address string `json:"address" note:"云服务器地址"`
	Port    int    `json:"port" note:"云服务器端口号"`
}
