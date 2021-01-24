package gcfg

type Config struct {
	Path   string `json:"-" note:"配置文件路径"`
	Module Module `json:"-" note:"模块信息"`

	Log     Log     `json:"log" note:"日志"`
	Svc     Svc     `json:"svc" note:"系统服务"`
	Cluster Cluster `json:"cluster" note:"集群配置"`
	Http    Http    `json:"http" note:"HTTP服务"`
	Https   Https   `json:"https" note:"HTTPS服务"`
	Proxy   string  `json:"proxy" note:"代理服务器IP地址（客户端不是来自代理服务器时，远程地址为当前连接地址）"`

	Site Site `json:"site" note:"站点配置"`
}
