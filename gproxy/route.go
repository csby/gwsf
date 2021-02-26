package gproxy

// 转发路由
type Route struct {
	// ture: the incoming TLS SNI server name is sni;
	// false: the incoming HTTP/1.x Host header name is httpHost
	IsTls bool

	// 监听地址，如"192.168.1.1:80", ":80"
	Address string

	// 转发域名，如"my.test.com", ""(全部转发)
	Domain string

	// 目标地址，如"172.16.100.85:8080"
	Target string

	// 版本号: 0-不添加头部；
	//1-添加代理头部（PROXY family srcIP srcPort targetIP targetPort）
	Version int
}
