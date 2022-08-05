package gcfg

type ServiceTomcat struct {
	Name        string   `json:"name" note:"项目名称"`
	ServiceName string   `json:"serviceName" note:"服务名称"`
	WebApp      string   `json:"webApp" note:"应用目录"`
	WebCfg      string   `json:"webCfg" note:"配置目录"`
	WebLog      string   `json:"webLog" note:"日志目录"`
	WebUrls     []string `json:"webUrls" note:"访问地址"`
}
